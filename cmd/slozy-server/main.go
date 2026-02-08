package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Middleware для логирования
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(rw, r)

		log.Printf("[%s] %s %s - %d %s",
			r.Method, r.URL.Path, r.RemoteAddr,
			rw.statusCode, time.Since(start))
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Структуры данных
type AlertingConfig struct {
	Enabled              bool                `json:"enabled"`
	NotificationChannels []string            `json:"notificationChannels"`
	BurnRateThresholds   []BurnRateThreshold `json:"burnRateThresholds"`
}

type BurnRateThreshold struct {
	Rate      float64 `json:"rate"`      // 10.0 (10x быстрее)
	Period    string  `json:"period"`    // "1h", "6h"
	Severity  string  `json:"severity"`  // "warning", "critical"
	AlertName string  `json:"alertName"` // "HighBurnRate"
}

type ServiceSLO struct {
	ServiceName        string            `json:"serviceName"`
	ServiceDescription string            `json:"serviceDescription"`
	SLOName            string            `json:"sloName"`
	SLODescription     string            `json:"sloDescription"`
	IndicatorType      string            `json:"indicatorType"`
	MetricSource       string            `json:"metricSource"`
	MetricQuery        string            `json:"metricQuery"`
	Target             float64           `json:"target"`
	Threshold          string            `json:"threshold"`
	BudgetingPeriod    string            `json:"budgetingPeriod"`
	Team               string            `json:"team"`
	Labels             map[string]string `json:"labels"`
	Alerting           AlertingConfig    `json:"alerting"`
}

// Структура для шаблона из YAML файла
type TemplateConfig struct {
	Name        string                 `yaml:"name"`
	DisplayName string                 `yaml:"displayName"`
	Description string                 `yaml:"description"`
	ServiceType string                 `yaml:"serviceType"`
	Defaults    map[string]interface{} `yaml:"defaults"`
	Metrics     map[string]string      `yaml:"metrics"`
	Examples    []struct {
		Name          string  `yaml:"name"`
		Description   string  `yaml:"description"`
		IndicatorType string  `yaml:"indicatorType,omitempty"`
		Target        float64 `yaml:"target,omitempty"`
		Threshold     string  `yaml:"threshold,omitempty"`
		MetricQuery   string  `yaml:"metricQuery"`
	} `yaml:"examples"`
	Alerting struct {
		Enabled              bool                `yaml:"enabled"`
		NotificationChannels []string            `yaml:"notificationChannels"`
		BurnRateThresholds   []BurnRateThreshold `yaml:"burnRateThresholds"`
	} `yaml:"alerting"`
}

// formatTarget форматирует target для красивого вывода
func formatTarget(target float64) string {
	if target == 100 {
		return "1"
	}

	commonValues := map[float64]string{
		99.9:   "0.999",
		99.95:  "0.9995",
		99.99:  "0.9999",
		99.999: "0.99999",
		99.5:   "0.995",
		99.0:   "0.99",
	}

	if str, ok := commonValues[target]; ok {
		return str
	}

	str := fmt.Sprintf("%.6f", target/100)
	str = strings.TrimRight(str, "0")
	str = strings.TrimRight(str, ".")
	return str
}

// Генерация OpenSLO YAML
func generateOpenSLO(slo ServiceSLO) ([]byte, error) {
	log.Printf("Generating SLO for service: %s, SLO: %s", slo.ServiceName, slo.SLOName)

	openslo := map[string]interface{}{
		"apiVersion": "openslo/v1alpha",
		"kind":       "SLO",
		"metadata": map[string]interface{}{
			"name":        fmt.Sprintf("%s-%s", slo.ServiceName, slo.SLOName),
			"displayName": slo.SLODescription,
		},
		"spec": map[string]interface{}{
			"description": slo.SLODescription,
			"service":     slo.ServiceName,
			"timeWindows": []map[string]interface{}{
				{
					"duration":  slo.BudgetingPeriod,
					"isRolling": true,
				},
			},
			"objectives": []map[string]interface{}{
				{
					"displayName": slo.SLODescription,
					"op":          "lte",
					"value":       formatTarget(slo.Target),
					"target":      "1",
				},
			},
		},
	}

	metadata := openslo["metadata"].(map[string]interface{})
	if slo.Labels == nil {
		slo.Labels = make(map[string]string)
	}
	slo.Labels["team"] = slo.Team
	slo.Labels["service"] = slo.ServiceName
	metadata["labels"] = slo.Labels

	spec := openslo["spec"].(map[string]interface{})
	switch strings.ToLower(slo.IndicatorType) {
	case "latency":
		spec["indicator"] = map[string]interface{}{
			"thresholdMetric": map[string]interface{}{
				"threshold": map[string]interface{}{
					"metricSourceRef": slo.MetricSource,
					"query":           slo.MetricQuery,
				},
			},
		}
	case "availability":
		spec["indicator"] = map[string]interface{}{
			"ratioMetric": map[string]interface{}{
				"counter": true,
				"good": map[string]interface{}{
					"metricSourceRef": slo.MetricSource,
					"query":           slo.MetricQuery + "_success",
				},
				"total": map[string]interface{}{
					"metricSourceRef": slo.MetricSource,
					"query":           slo.MetricQuery + "_total",
				},
			},
		}
	default:
		spec["indicator"] = map[string]interface{}{
			"ratioMetric": map[string]interface{}{
				"counter": true,
				"good": map[string]interface{}{
					"metricSourceRef": slo.MetricSource,
					"query":           slo.MetricQuery,
				},
				"total": map[string]interface{}{
					"metricSourceRef": slo.MetricSource,
					"query":           slo.MetricQuery,
				},
			},
		}
	}

	// Добавляем Alert Policies если включены
	if slo.Alerting.Enabled && len(slo.Alerting.BurnRateThresholds) > 0 {
		var alertPolicies []map[string]interface{}

		for _, br := range slo.Alerting.BurnRateThresholds {
			alertName := br.AlertName
			if alertName == "" {
				alertName = fmt.Sprintf("HighBurnRate_%sx_%s",
					strings.Replace(fmt.Sprintf("%.1f", br.Rate), ".", "_", -1),
					strings.Replace(br.Period, "m", "min", -1))
			}

			alertPolicy := map[string]interface{}{
				"condition": map[string]interface{}{
					"kind":           "burnrate",
					"threshold":      br.Rate,
					"lookbackWindow": br.Period,
				},
				"notification": map[string]interface{}{
					"channels": slo.Alerting.NotificationChannels,
					"telegram": map[string]interface{}{
						"parse_mode":            "Markdown",
						"disable_notifications": false,
					},
				},
				"labels": map[string]string{
					"severity":   br.Severity,
					"service":    slo.ServiceName,
					"slo":        slo.SLOName,
					"team":       slo.Team,
					"alert_name": alertName,
				},
				"annotations": map[string]string{
					"summary": fmt.Sprintf("High error budget burn rate for %s", slo.SLOName),
					"description": fmt.Sprintf("%s is burning error budget %.1f times faster than allowed for %s",
						slo.ServiceName, br.Rate, br.Period),
					"dashboard": fmt.Sprintf("https://grafana.example.com/d/%s-%s", slo.ServiceName, slo.SLOName),
				},
			}
			alertPolicies = append(alertPolicies, alertPolicy)
		}

		spec["alertPolicies"] = alertPolicies
	}

	return yaml.Marshal(openslo)
}

// generatePrometheusAlertRules генерирует Prometheus Alert Rules
func generatePrometheusAlertRules(slo ServiceSLO) ([]byte, error) {
	if !slo.Alerting.Enabled || len(slo.Alerting.BurnRateThresholds) == 0 {
		return nil, nil
	}

	targetValue := formatTarget(slo.Target)

	rules := map[string]interface{}{
		"groups": []map[string]interface{}{
			{
				"name":  fmt.Sprintf("slo-%s-%s", slo.ServiceName, slo.SLOName),
				"rules": []map[string]interface{}{},
			},
		},
	}

	group := rules["groups"].([]map[string]interface{})[0]

	for _, br := range slo.Alerting.BurnRateThresholds {
		expr := ""
		if slo.IndicatorType == "latency" {
			expr = fmt.Sprintf(`(
                (1 - (clamp_min(%s, 0) / %s))
                / (1 - %s)
            ) > %.1f`,
				slo.MetricQuery,
				slo.Threshold,
				targetValue,
				br.Rate)
		} else if slo.IndicatorType == "availability" {
			expr = fmt.Sprintf(`(
                (1 - (clamp_min(%s, 0)))
                / (1 - %s)
            ) > %.1f`,
				slo.MetricQuery,
				targetValue,
				br.Rate)
		} else {
			expr = fmt.Sprintf(`(
                (1 - (clamp_min(%s, 0)))
                / (1 - %s)
            ) > %.1f`,
				slo.MetricQuery,
				targetValue,
				br.Rate)
		}

		alertName := br.AlertName
		if alertName == "" {
			alertName = fmt.Sprintf("%s_%s_BurnRateHigh",
				strings.Replace(slo.ServiceName, "-", "_", -1),
				strings.Replace(slo.SLOName, "-", "_", -1))
		}

		// Создаем метки с информацией о каналах уведомлений
		labels := map[string]string{
			"severity":            br.Severity,
			"service":             slo.ServiceName,
			"slo":                 slo.SLOName,
			"slo_indicator":       slo.IndicatorType,
			"slo_target":          fmt.Sprintf("%.1f", slo.Target),
			"team":                slo.Team,
			"burn_rate_threshold": fmt.Sprintf("%.1fx", br.Rate),
		}

		// Добавляем информацию о каналах уведомлений если они есть
		if len(slo.Alerting.NotificationChannels) > 0 {
			labels["notification_channels"] = strings.Join(slo.Alerting.NotificationChannels, ",")
		}

		alertRule := map[string]interface{}{
			"alert":  alertName,
			"expr":   expr,
			"for":    br.Period,
			"labels": labels,
			"annotations": map[string]string{
				"summary":     fmt.Sprintf("High error budget burn rate for %s/%s", slo.ServiceName, slo.SLOName),
				"description": fmt.Sprintf("{{ $labels.service }} {{ $labels.slo }} is burning error budget {{ $labels.burn_rate_threshold }} faster than allowed.\nCurrent burn rate: {{ $value }}x\nSLO Target: {{ $labels.slo_target }}%%\nFor: %s", br.Period),
				"dashboard":   fmt.Sprintf("https://grafana.example.com/d/%s-%s", slo.ServiceName, slo.SLOName),
			},
		}

		group["rules"] = append(group["rules"].([]map[string]interface{}), alertRule)

		if br.Severity == "warning" && br.Rate < 100 {
			criticalAlertRule := map[string]interface{}{
				"alert": alertName + "_Critical",
				"expr":  fmt.Sprintf(`%s > %.1f`, strings.TrimSuffix(expr, fmt.Sprintf("> %.1f", br.Rate)), br.Rate*5),
				"for":   "5m",
				"labels": map[string]string{
					"severity":            "critical",
					"service":             slo.ServiceName,
					"slo":                 slo.SLOName,
					"slo_indicator":       slo.IndicatorType,
					"slo_target":          fmt.Sprintf("%.1f", slo.Target),
					"team":                slo.Team,
					"burn_rate_threshold": fmt.Sprintf("%.1fx", br.Rate*5),
				},
				"annotations": map[string]string{
					"summary":     fmt.Sprintf("CRITICAL: Very high error budget burn rate for %s/%s", slo.ServiceName, slo.SLOName),
					"description": "Error budget is being consumed extremely fast. Immediate action required!",
				},
			}
			group["rules"] = append(group["rules"].([]map[string]interface{}), criticalAlertRule)
		}
	}

	return yaml.Marshal(rules)
}

// generateAlertManagerConfig генерирует конфиг для Alertmanager
func generateAlertManagerConfig(slo ServiceSLO) ([]byte, error) {
	if !slo.Alerting.Enabled {
		return nil, nil
	}

	// Получаем Telegram токен из переменной окружения
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	// Базовые receivers
	receivers := []map[string]interface{}{
		{
			"name": "slo-notifications",
		},
		{
			"name": "critical-alerts",
		},
	}

	// Создаем конфиги для разных типов уведомлений
	slackConfigs := []map[string]interface{}{}
	emailConfigs := []map[string]interface{}{}
	pagerdutyConfigs := []map[string]interface{}{}
	telegramConfigs := []map[string]interface{}{}
	webhookConfigs := []map[string]interface{}{}

	// Парсим notification channels из SLO
	for _, channel := range slo.Alerting.NotificationChannels {
		channel = strings.TrimSpace(channel)

		if strings.HasPrefix(channel, "slack") {
			// slack#channel-name или slack:webhook-url
			var slackConfig map[string]interface{}
			if strings.Contains(channel, "#") {
				parts := strings.Split(channel, "#")
				channelName := parts[1]
				slackConfig = map[string]interface{}{
					"api_url":  "https://hooks.slack.com/services/T00000000/B00000000/secret",
					"channel":  "#" + channelName,
					"title":    "{{ .GroupLabels.service }} SLO Alert",
					"text":     "{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}",
					"icon_url": "https://raw.githubusercontent.com/philyuchkoff/slozy/main/static/favicon.ico",
				}
			} else {
				slackConfig = map[string]interface{}{
					"api_url":  channel,
					"channel":  "#slo-alerts",
					"title":    "{{ .GroupLabels.service }} SLO Alert",
					"text":     "{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}",
					"icon_url": "https://raw.githubusercontent.com/philyuchkoff/slozy/main/static/favicon.ico",
				}
			}
			slackConfigs = append(slackConfigs, slackConfig)

		} else if strings.HasPrefix(channel, "email") {
			// email:address@example.com
			parts := strings.Split(channel, ":")
			emailTo := "sre-team@example.com"
			if len(parts) > 1 {
				emailTo = parts[1]
			}
			emailConfigs = append(emailConfigs, map[string]interface{}{
				"to":        emailTo,
				"from":      "slo-alerts@example.com",
				"smarthost": "localhost:25",
				"subject":   "SLO Alert: {{ .GroupLabels.service }} {{ .GroupLabels.slo }}",
				"html":      "{{ range .Alerts }}<p><strong>{{ .Annotations.summary }}</strong><br>{{ .Annotations.description }}</p>{{ end }}",
			})

		} else if strings.HasPrefix(channel, "pagerduty") {
			// pagerduty или pagerduty:service-key
			parts := strings.Split(channel, ":")
			serviceKey := "your-pagerduty-key"
			if len(parts) > 1 {
				serviceKey = parts[1]
			}
			pagerdutyConfigs = append(pagerdutyConfigs, map[string]interface{}{
				"service_key": serviceKey,
				"description": "{{ .GroupLabels.alertname }}",
				"severity":    "{{ .GroupLabels.severity }}",
			})

		} else if strings.HasPrefix(channel, "telegram") {
			// telegram:chat-id
			if telegramBotToken != "" {
				parts := strings.Split(channel, ":")
				chatID := telegramChatID
				if len(parts) > 1 && parts[1] != "" {
					chatID = parts[1]
				}

				if chatID != "" {
					telegramConfig := map[string]interface{}{
						"api_url":               fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramBotToken),
						"chat_id":               chatID,
						"message":               "🚨 *{{ .GroupLabels.severity | toUpper }} SLO Alert*\n\n*Service:* {{ .GroupLabels.service }}\n*SLO:* {{ .GroupLabels.slo }}\n*Severity:* {{ .GroupLabels.severity }}\n*Alert:* {{ .GroupLabels.alertname }}\n\n{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}\n\n{{ if (index .Alerts 0).Annotations.dashboard }}[View Dashboard]({{ (index .Alerts 0).Annotations.dashboard }}){{ end }}",
						"parse_mode":            "Markdown",
						"disable_notifications": false,
						"send_resolved":         true,
					}
					telegramConfigs = append(telegramConfigs, telegramConfig)
				}
			}

		} else if strings.HasPrefix(channel, "webhook") {
			// webhook:url
			parts := strings.Split(channel, ":")
			url := "http://localhost:9093/webhook"
			if len(parts) > 1 {
				url = parts[1]
			}
			webhookConfigs = append(webhookConfigs, map[string]interface{}{
				"url":           url,
				"send_resolved": true,
			})

		}
	}

	// Добавляем конфиги в receivers
	if len(slackConfigs) > 0 {
		receivers[0]["slack_configs"] = slackConfigs
	}
	if len(emailConfigs) > 0 {
		receivers[0]["email_configs"] = emailConfigs
	}
	if len(pagerdutyConfigs) > 0 {
		receivers[1]["pagerduty_configs"] = pagerdutyConfigs
	}
	if len(telegramConfigs) > 0 {
		receivers[0]["telegram_configs"] = telegramConfigs
		// Также добавляем Telegram в critical-alerts
		receivers[1]["telegram_configs"] = telegramConfigs
	}
	if len(webhookConfigs) > 0 {
		receivers[0]["webhook_configs"] = webhookConfigs
	}

	// Если Telegram токен есть, но не указан в channels, добавляем по умолчанию
	if telegramBotToken != "" && telegramChatID != "" && len(telegramConfigs) == 0 {
		defaultTelegramConfig := map[string]interface{}{
			"api_url":               fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramBotToken),
			"chat_id":               telegramChatID,
			"message":               "🚨 *{{ .GroupLabels.severity | toUpper }} SLO Alert*\n\n*Service:* {{ .GroupLabels.service }}\n*SLO:* {{ .GroupLabels.slo }}\n*Severity:* {{ .GroupLabels.severity }}\n\n{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}",
			"parse_mode":            "Markdown",
			"disable_notifications": false,
			"send_resolved":         true,
		}
		receivers[0]["telegram_configs"] = []map[string]interface{}{defaultTelegramConfig}
		receivers[1]["telegram_configs"] = []map[string]interface{}{defaultTelegramConfig}
	}

	config := map[string]interface{}{
		"global": map[string]interface{}{
			"smtp_smarthost": "localhost:25",
			"smtp_from":      "slo-alerts@example.com",
		},
		"route": map[string]interface{}{
			"group_by":        []string{"alertname", "service", "slo", "severity"},
			"group_wait":      "30s",
			"group_interval":  "5m",
			"repeat_interval": "12h",
			"receiver":        "slo-notifications",
			"routes": []map[string]interface{}{
				{
					"match": map[string]string{
						"severity": "critical",
					},
					"receiver": "critical-alerts",
				},
				{
					"match": map[string]string{
						"service": slo.ServiceName,
						"slo":     slo.SLOName,
					},
					"receiver": "slo-notifications",
				},
			},
		},
		"receivers": receivers,
		"inhibit_rules": []map[string]interface{}{
			{
				"source_match": map[string]string{
					"severity": "critical",
				},
				"target_match": map[string]string{
					"severity": "warning",
				},
				"equal": []string{"service", "slo"},
			},
		},
	}

	return yaml.Marshal(config)
}

// generateGrafanaDashboard генерирует дашборд Grafana с алертами
func generateGrafanaDashboard(slo ServiceSLO) ([]byte, error) {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"title":    fmt.Sprintf("%s - %s SLO", slo.ServiceName, slo.SLOName),
			"tags":     []string{"slo", slo.ServiceName, slo.Team, slo.IndicatorType},
			"timezone": "browser",
			"panels": []map[string]interface{}{
				{
					"title": "SLO Burn Rate",
					"type":  "graph",
					"gridPos": map[string]interface{}{
						"h": 8, "w": 12, "x": 0, "y": 0,
					},
					"targets": []map[string]interface{}{
						{
							"expr": fmt.Sprintf(`(
                                (1 - (clamp_min(%s, 0)))
                                / (1 - %s)
                            )`, slo.MetricQuery, formatTarget(slo.Target)),
							"legendFormat": "Burn Rate",
						},
					},
					"alert": map[string]interface{}{
						"conditions": []map[string]interface{}{
							{
								"evaluator": map[string]interface{}{
									"params": []float64{10.0},
									"type":   "gt",
								},
								"operator": map[string]interface{}{
									"type": "and",
								},
								"query": map[string]interface{}{
									"params": []string{"A", "5m", "now"},
								},
								"reducer": map[string]interface{}{
									"params": []string{},
									"type":   "avg",
								},
								"type": "query",
							},
						},
						"executionErrorState": "alerting",
						"for":                 "1h",
						"frequency":           "1m",
						"handler":             1,
						"name":                "High Burn Rate Alert",
						"noDataState":         "no_data",
						"notifications": []map[string]interface{}{
							{
								"id": 1,
							},
						},
					},
				},
				{
					"title": "Error Budget Remaining",
					"type":  "stat",
					"gridPos": map[string]interface{}{
						"h": 4, "w": 6, "x": 0, "y": 8,
					},
					"targets": []map[string]interface{}{
						{
							"expr": fmt.Sprintf(`100 - (
                                (1 - (clamp_min(%s, 0)))
                                / (1 - %s)
                                * 100
                            )`, slo.MetricQuery, formatTarget(slo.Target)),
							"legendFormat": "Budget %",
						},
					},
					"options": map[string]interface{}{
						"fieldOptions": map[string]interface{}{
							"calcs": []string{"lastNotNull"},
							"defaults": map[string]interface{}{
								"unit": "percent",
								"min":  0,
								"max":  100,
								"thresholds": map[string]interface{}{
									"steps": []map[string]interface{}{
										{"color": "red", "value": nil},
										{"color": "yellow", "value": 20},
										{"color": "green", "value": 50},
									},
								},
							},
						},
					},
				},
				{
					"title": "Alert History",
					"type":  "alertlist",
					"gridPos": map[string]interface{}{
						"h": 8, "w": 12, "x": 12, "y": 0,
					},
					"options": map[string]interface{}{
						"show":        "current",
						"sortOrder":   1,
						"stateFilter": []string{"alerting", "pending"},
					},
				},
			},
			"templating": map[string]interface{}{
				"list": []map[string]interface{}{
					{
						"name":  "service",
						"query": slo.ServiceName,
						"type":  "constant",
						"hide":  2,
					},
					{
						"name":  "slo",
						"query": slo.SLOName,
						"type":  "constant",
						"hide":  2,
					},
				},
			},
			"refresh":       "30s",
			"schemaVersion": 27,
			"version":       1,
			"uid":           fmt.Sprintf("%s-%s-slo", strings.Replace(slo.ServiceName, " ", "-", -1), strings.Replace(slo.SLOName, " ", "-", -1)),
		},
		"folderId":  0,
		"overwrite": true,
	}

	return json.MarshalIndent(dashboard, "", "  ")
}

// generateMultiBurnRateDashboard создает дашборд с несколькими burn rate thresholds
func generateMultiBurnRateDashboard(slo ServiceSLO) ([]byte, error) {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"title":  fmt.Sprintf("%s - %s Multi Burn Rate", slo.ServiceName, slo.SLOName),
			"tags":   []string{"slo", "burnrate", slo.ServiceName},
			"panels": []map[string]interface{}{},
			"templating": map[string]interface{}{
				"list": []map[string]interface{}{
					{
						"name":       "burn_rate_window",
						"label":      "Burn Rate Window",
						"query":      "1h,6h,1d,3d",
						"type":       "custom",
						"multi":      false,
						"includeAll": false,
						"current": map[string]interface{}{
							"text":  "1h",
							"value": "1h",
						},
					},
				},
			},
		},
	}

	panels := dashboard["dashboard"].(map[string]interface{})["panels"].([]map[string]interface{})

	for i, br := range slo.Alerting.BurnRateThresholds {
		panel := map[string]interface{}{
			"title": fmt.Sprintf("%s Burn Rate (%s)", strings.Title(br.Severity), br.Period),
			"type":  "graph",
			"gridPos": map[string]interface{}{
				"h": 8, "w": 12, "x": (i % 2) * 12, "y": int(math.Floor(float64(i)/2)) * 8,
			},
			"targets": []map[string]interface{}{
				{
					"expr": fmt.Sprintf(`(
                        (1 - (clamp_min(%s, 0)))
                        / (1 - %s)
                    )`, slo.MetricQuery, formatTarget(slo.Target)),
					"legendFormat": "Current",
				},
				{
					"expr":         fmt.Sprintf("%.1f", br.Rate),
					"legendFormat": fmt.Sprintf("%s Threshold (%.1fx)", strings.Title(br.Severity), br.Rate),
				},
			},
		}
		panels = append(panels, panel)
	}

	return json.MarshalIndent(dashboard, "", "  ")
}

// generateReadme генерирует README файл с инструкциями
func generateReadme(slo ServiceSLO, mainFilename string, additionalFiles map[string]string) string {
	var readme strings.Builder

	readme.WriteString(fmt.Sprintf("# SLO Configuration: %s - %s\n\n", slo.ServiceName, slo.SLOName))
	readme.WriteString("## Service Information\n")
	readme.WriteString(fmt.Sprintf("- **Service**: %s\n", slo.ServiceName))
	readme.WriteString(fmt.Sprintf("- **SLO**: %s\n", slo.SLOName))
	readme.WriteString(fmt.Sprintf("- **Target**: %.1f%%\n", slo.Target))
	readme.WriteString(fmt.Sprintf("- **Indicator**: %s\n", slo.IndicatorType))
	readme.WriteString(fmt.Sprintf("- **Team**: %s\n", slo.Team))
	readme.WriteString(fmt.Sprintf("- **Created**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	readme.WriteString("## Files Generated\n\n")
	readme.WriteString(fmt.Sprintf("### 1. Main SLO Definition\n**File**: %s\n", mainFilename))
	readme.WriteString("Contains the OpenSLO compliant SLO definition with objectives and alert policies.\n\n")

	if len(additionalFiles) > 0 {
		readme.WriteString("### 2. Alerting Configuration\n")

		if promFile, ok := additionalFiles["prometheusAlerts"]; ok {
			readme.WriteString(fmt.Sprintf("\n#### Prometheus Alert Rules\n**File**: %s\n", promFile))
			readme.WriteString("Contains Prometheus alert rules for monitoring error budget burn rate.\n")
		}

		if amFile, ok := additionalFiles["alertManagerConfig"]; ok {
			readme.WriteString(fmt.Sprintf("\n#### Alertmanager Configuration\n**File**: %s\n", amFile))
			readme.WriteString("Sample Alertmanager configuration for routing SLO alerts.\n")
		}

		if grafanaFile, ok := additionalFiles["grafanaDashboard"]; ok {
			readme.WriteString(fmt.Sprintf("\n#### Grafana Dashboard\n**File**: %s\n", grafanaFile))
			readme.WriteString("Complete Grafana dashboard for SLO monitoring.\n")
		}
	}

	if slo.Alerting.Enabled && len(slo.Alerting.BurnRateThresholds) > 0 {
		readme.WriteString("\n## Alerting Strategy\n\n")
		readme.WriteString("### Burn Rate Thresholds\n")

		for _, br := range slo.Alerting.BurnRateThresholds {
			action := "Notify team, investigate within 24h"
			if br.Severity == "critical" {
				action = "Immediate action required, page on-call"
			}

			readme.WriteString(fmt.Sprintf("\n- **%s**: > %.1fx for %s\n", br.AlertName, br.Rate, br.Period))
			readme.WriteString(fmt.Sprintf("  - Severity: %s\n", br.Severity))
			readme.WriteString(fmt.Sprintf("  - Action: %s\n", action))
		}

		if len(slo.Alerting.NotificationChannels) > 0 {
			readme.WriteString("\n### Notification Channels\n")
			for _, channel := range slo.Alerting.NotificationChannels {
				readme.WriteString(fmt.Sprintf("- %s\n", channel))
			}

			// Добавляем инструкции для Telegram
			for _, channel := range slo.Alerting.NotificationChannels {
				if strings.HasPrefix(channel, "telegram") {
					readme.WriteString("\n### Telegram Setup Instructions\n")
					readme.WriteString("To configure Telegram notifications:\n")
					readme.WriteString("1. Set environment variables on your Alertmanager server:\n")
					readme.WriteString("   ```bash\n")
					readme.WriteString("   export TELEGRAM_BOT_TOKEN='your_bot_token_from_BotFather'\n")
					readme.WriteString("   export TELEGRAM_CHAT_ID='your_chat_id'\n")
					readme.WriteString("   ```\n")
					readme.WriteString("2. Restart Alertmanager with these environment variables\n")
					readme.WriteString("3. Ensure Alertmanager has internet access to api.telegram.org\n")
					break
				}
			}
		}
	}

	readme.WriteString(fmt.Sprintf("\n## Error Budget\n\n"))
	readme.WriteString(fmt.Sprintf("Error Budget: %.3f%% allowable failures over %s\n", 100-slo.Target, slo.BudgetingPeriod))

	readme.WriteString("\n## Useful Links\n")
	readme.WriteString("- [OpenSLO Specification](https://openslo.com)\n")
	readme.WriteString("- [SLOzy GitHub](https://github.com/philyuchkoff/slozy)\n")
	readme.WriteString("- [Alertmanager Documentation](https://prometheus.io/docs/alerting/latest/configuration/)\n")
	readme.WriteString("- [Telegram Bot API](https://core.telegram.org/bots/api)\n")

	return readme.String()
}

// Валидация SLO
func validateSLO(slo ServiceSLO) []string {
	var errors []string

	if strings.TrimSpace(slo.ServiceName) == "" {
		errors = append(errors, "Service name is required")
	}

	if strings.TrimSpace(slo.SLOName) == "" {
		errors = append(errors, "SLO name is required")
	}

	if slo.Target < 90 || slo.Target > 100 {
		errors = append(errors, "Target must be between 90 and 100")
	}

	if strings.TrimSpace(slo.MetricQuery) == "" {
		errors = append(errors, "Metric query is required")
	}

	return errors
}

// createZipArchive создает ZIP архив со всеми файлами
func createZipArchive(files map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create file in zip: %w", err)
		}

		if _, err := writer.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write file to zip: %w", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Обработчик создания SLO с ZIP архивом
func createSLOHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received SLO creation request")

	if r.Method != http.MethodPost {
		log.Printf("Wrong method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	var slo ServiceSLO
	if err := json.NewDecoder(r.Body).Decode(&slo); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Parsed SLO data: Service: %s, SLO: %s", slo.ServiceName, slo.SLOName)

	if validationErrors := validateSLO(slo); len(validationErrors) > 0 {
		log.Printf("Validation failed: %v", validationErrors)
		response := map[string]interface{}{
			"success": false,
			"errors":  validationErrors,
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(response)
		return
	}

	if slo.BudgetingPeriod == "" {
		slo.BudgetingPeriod = "28d"
	}
	if slo.IndicatorType == "" {
		slo.IndicatorType = "latency"
	}
	if slo.MetricSource == "" {
		slo.MetricSource = "prometheus"
	}
	if slo.Threshold == "" {
		slo.Threshold = "200ms"
	}
	if slo.Team == "" {
		slo.Team = "unknown"
	}

	// Создаем временную директорию для файлов
	tempDir, err := os.MkdirTemp("", "slozy-*")
	if err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create temp directory: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir) // Удаляем временную директорию при выходе

	// Генерируем основной SLO YAML
	yamlContent, err := generateOpenSLO(slo)
	if err != nil {
		log.Printf("YAML generation error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate YAML: %v", err), http.StatusInternalServerError)
		return
	}

	safeServiceName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '_'
	}, slo.ServiceName)

	safeSLOName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '_'
	}, slo.SLOName)

	baseFilename := fmt.Sprintf("%s-%s", safeServiceName, safeSLOName)
	mainFilename := baseFilename + ".yaml"
	mainFilePath := filepath.Join(tempDir, mainFilename)

	log.Printf("Saving main SLO to: %s", mainFilePath)
	if err := os.WriteFile(mainFilePath, yamlContent, 0644); err != nil {
		log.Printf("File write error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	// Собираем все файлы для ZIP архива
	allFiles := map[string][]byte{
		mainFilename: yamlContent,
	}

	// Генерируем дополнительные файлы
	if slo.Alerting.Enabled {
		// 1. Prometheus Alert Rules
		prometheusRules, err := generatePrometheusAlertRules(slo)
		if err == nil && prometheusRules != nil {
			promFilename := baseFilename + "-prometheus-alerts.yaml"
			promFilePath := filepath.Join(tempDir, promFilename)

			if err := os.WriteFile(promFilePath, prometheusRules, 0644); err == nil {
				log.Printf("Prometheus alert rules saved: %s", promFilename)
				allFiles[promFilename] = prometheusRules
			}
		}

		// 2. Alertmanager Config
		alertManagerConfig, err := generateAlertManagerConfig(slo)
		if err == nil && alertManagerConfig != nil {
			amFilename := baseFilename + "-alertmanager-config.yaml"
			amFilePath := filepath.Join(tempDir, amFilename)

			if err := os.WriteFile(amFilePath, alertManagerConfig, 0644); err == nil {
				log.Printf("Alertmanager config saved: %s", amFilename)
				allFiles[amFilename] = alertManagerConfig
			}
		}

		// 3. Grafana Dashboard
		grafanaDashboard, err := generateGrafanaDashboard(slo)
		if err == nil && grafanaDashboard != nil {
			grafanaFilename := baseFilename + "-grafana-dashboard.json"
			grafanaFilePath := filepath.Join(tempDir, grafanaFilename)

			if err := os.WriteFile(grafanaFilePath, grafanaDashboard, 0644); err == nil {
				log.Printf("Grafana dashboard saved: %s", grafanaFilename)
				allFiles[grafanaFilename] = grafanaDashboard
			}
		}

		// 4. Multi Burn Rate Dashboard
		multiBurnRateDashboard, err := generateMultiBurnRateDashboard(slo)
		if err == nil && multiBurnRateDashboard != nil {
			multiFilename := baseFilename + "-multi-burnrate-dashboard.json"
			multiFilePath := filepath.Join(tempDir, multiFilename)

			if err := os.WriteFile(multiFilePath, multiBurnRateDashboard, 0644); err == nil {
				log.Printf("Multi burn rate dashboard saved: %s", multiFilename)
				allFiles[multiFilename] = multiBurnRateDashboard
			}
		}

		// 5. README
		additionalFiles := map[string]string{
			"prometheusAlerts":   baseFilename + "-prometheus-alerts.yaml",
			"alertManagerConfig": baseFilename + "-alertmanager-config.yaml",
			"grafanaDashboard":   baseFilename + "-grafana-dashboard.json",
		}
		readmeContent := generateReadme(slo, mainFilename, additionalFiles)
		readmeFilename := baseFilename + "-README.md"
		readmeFilePath := filepath.Join(tempDir, readmeFilename)

		if err := os.WriteFile(readmeFilePath, []byte(readmeContent), 0644); err == nil {
			log.Printf("README saved: %s", readmeFilename)
			allFiles[readmeFilename] = []byte(readmeContent)
		}
	}

	// Создаем ZIP архив
	log.Printf("Creating ZIP archive with %d files", len(allFiles))
	zipBytes, err := createZipArchive(allFiles)
	if err != nil {
		log.Printf("Failed to create ZIP archive: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create ZIP archive: %v", err), http.StatusInternalServerError)
		return
	}

	// Генерируем уникальное имя для ZIP файла
	timestamp := time.Now().Format("20060102-150405")
	zipFilename := fmt.Sprintf("%s-%s.zip", baseFilename, timestamp)

	// Отправляем ZIP архив клиенту
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipFilename))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

	if _, err := w.Write(zipBytes); err != nil {
		log.Printf("Failed to send ZIP archive: %v", err)
	}

	log.Printf("ZIP archive sent successfully: %s (%d bytes)", zipFilename, len(zipBytes))
}

// Функция для загрузки шаблонов из YAML файлов
func loadTemplatesFromFiles() ([]map[string]interface{}, error) {
	templatesDir := "./templates"
	var templates []map[string]interface{}

	// Создаем директорию если не существует
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Читаем файлы в директории
	files, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			filePath := filepath.Join(templatesDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Warning: Failed to read template file %s: %v", filePath, err)
				continue
			}

			var template TemplateConfig
			if err := yaml.Unmarshal(content, &template); err != nil {
				log.Printf("Warning: Failed to parse template file %s: %v", filePath, err)
				continue
			}

			// Конвертируем в формат, ожидаемый фронтендом
			frontendTemplate := map[string]interface{}{
				"name":        template.Name,
				"displayName": template.DisplayName,
				"description": template.Description,
				"serviceType": template.ServiceType,
				"defaults":    template.Defaults,
			}

			// Добавляем алертинг если есть
			if template.Alerting.Enabled {
				if frontendTemplate["defaults"] == nil {
					frontendTemplate["defaults"] = make(map[string]interface{})
				}
				frontendTemplate["defaults"].(map[string]interface{})["alerting"] = map[string]interface{}{
					"enabled":              template.Alerting.Enabled,
					"notificationChannels": template.Alerting.NotificationChannels,
					"burnRateThresholds":   template.Alerting.BurnRateThresholds,
				}
			}

			// Добавляем примеры метрик если есть
			if len(template.Examples) > 0 {
				// Используем первый пример как дефолтный metricQuery
				if frontendTemplate["defaults"] == nil {
					frontendTemplate["defaults"] = make(map[string]interface{})
				}
				frontendTemplate["defaults"].(map[string]interface{})["metricQuery"] = template.Examples[0].MetricQuery

				// Добавляем остальные параметры из первого примера если они есть
				if template.Examples[0].IndicatorType != "" {
					frontendTemplate["defaults"].(map[string]interface{})["indicatorType"] = template.Examples[0].IndicatorType
				}
				if template.Examples[0].Target > 0 {
					frontendTemplate["defaults"].(map[string]interface{})["target"] = template.Examples[0].Target
				}
				if template.Examples[0].Threshold != "" {
					frontendTemplate["defaults"].(map[string]interface{})["threshold"] = template.Examples[0].Threshold
				}
			}

			// Добавляем serviceType в defaults для фронтенда
			if frontendTemplate["defaults"] == nil {
				frontendTemplate["defaults"] = make(map[string]interface{})
			}
			frontendTemplate["defaults"].(map[string]interface{})["serviceType"] = template.ServiceType

			templates = append(templates, frontendTemplate)
		}
	}

	// Если нет файлов, возвращаем дефолтные шаблоны
	if len(templates) == 0 {
		return getDefaultTemplates(), nil
	}

	return templates, nil
}

// Функция для получения дефолтных шаблонов
func getDefaultTemplates() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "web-api-latency",
			"displayName": "Web API Latency",
			"description": "Latency SLO for web APIs with alerting",
			"serviceType": "web",
			"defaults": map[string]interface{}{
				"indicatorType":   "latency",
				"target":          99.9,
				"threshold":       "200ms",
				"budgetingPeriod": "28d",
				"metricSource":    "prometheus",
				"metricQuery":     "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job=\"$SERVICE\"}[5m]))",
				"alerting": map[string]interface{}{
					"enabled":              true,
					"notificationChannels": []string{"slack#slo-alerts", "email:sre@example.com", "telegram:your_chat_id"},
					"burnRateThresholds": []map[string]interface{}{
						{
							"rate":      10.0,
							"period":    "1h",
							"severity":  "warning",
							"alertName": "HighBurnRate",
						},
						{
							"rate":      50.0,
							"period":    "5m",
							"severity":  "critical",
							"alertName": "CriticalBurnRate",
						},
					},
				},
			},
		},
		{
			"name":        "web-api-availability",
			"displayName": "Web API Availability",
			"description": "Availability SLO for web APIs with alerting",
			"serviceType": "web",
			"defaults": map[string]interface{}{
				"indicatorType":   "availability",
				"target":          99.95,
				"threshold":       "0.0005",
				"budgetingPeriod": "28d",
				"metricSource":    "prometheus",
				"metricQuery":     "sum(rate(http_requests_total{job=\"$SERVICE\", status=~\"2..\"}[5m])) / sum(rate(http_requests_total{job=\"$SERVICE\"}[5m]))",
				"alerting": map[string]interface{}{
					"enabled":              true,
					"notificationChannels": []string{"slack#slo-alerts", "telegram:your_chat_id"},
					"burnRateThresholds": []map[string]interface{}{
						{
							"rate":     5.0,
							"period":   "1h",
							"severity": "warning",
						},
					},
				},
			},
		},
		{
			"name":        "database-latency",
			"displayName": "Database Query Latency",
			"description": "Latency SLO for database queries",
			"serviceType": "database",
			"defaults": map[string]interface{}{
				"indicatorType":   "latency",
				"target":          99.9,
				"threshold":       "100ms",
				"budgetingPeriod": "28d",
				"metricSource":    "prometheus",
				"metricQuery":     "histogram_quantile(0.99, rate(pg_query_duration_seconds_bucket{database=\"$DB\"}[5m]))",
				"alerting": map[string]interface{}{
					"enabled":              true,
					"notificationChannels": []string{"pagerduty", "telegram:your_chat_id"},
					"burnRateThresholds": []map[string]interface{}{
						{
							"rate":     20.0,
							"period":   "30m",
							"severity": "critical",
						},
					},
				},
			},
		},
	}
}

// Обработчик шаблонов
func templatesHandler(w http.ResponseWriter, r *http.Request) {
	templates, err := loadTemplatesFromFiles()
	if err != nil {
		log.Printf("Warning: Failed to load templates from files, using defaults: %v", err)
		templates = getDefaultTemplates()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{"templates": templates})
}

// Обработчик для создания дефолтных шаблонов
func createDefaultTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := copyExampleTemplates(); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Default templates created successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// Функция для копирования примеров шаблонов
func copyExampleTemplates() error {
	templatesDir := "./templates"

	// Проверяем, есть ли уже файлы
	files, err := os.ReadDir(templatesDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	if len(files) > 0 {
		return fmt.Errorf("templates directory is not empty")
	}

	// Создаем примеры шаблонов
	exampleTemplates := map[string]string{
		"web-api-latency.yaml": `name: web-api-latency
displayName: Web API Latency
description: Latency SLO for web APIs with alerting
serviceType: web

defaults:
  indicatorType: latency
  metricSource: prometheus
  budgetingPeriod: 28d
  target: 99.9
  threshold: 200ms

metrics:
  latency_99: |
    histogram_quantile(0.99,
      rate(http_request_duration_seconds_bucket{job="$SERVICE"}[5m])
    )
  latency_95: |
    histogram_quantile(0.95,
      rate(http_request_duration_seconds_bucket{job="$SERVICE"}[5m])
    )
  request_rate: |
    rate(http_requests_total{job="$SERVICE"}[5m])

examples:
  - name: API Latency 99th percentile
    description: 99th percentile response time
    metricQuery: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="my-service"}[5m]))
  
  - name: API Success Rate
    description: HTTP success rate
    indicatorType: availability
    target: 99.95
    metricQuery: sum(rate(http_requests_total{job="my-service", status=~"2.."}[5m])) / sum(rate(http_requests_total{job="my-service"}[5m]))

alerting:
  enabled: true
  notificationChannels: ["slack#slo-alerts", "email:sre@example.com", "telegram:your_chat_id"]
  burnRateThresholds:
    - rate: 10.0
      period: 1h
      severity: warning
      alertName: HighBurnRate
    - rate: 50.0
      period: 5m
      severity: critical
      alertName: CriticalBurnRate`,

		"database.yaml": `name: database
displayName: Database Service
description: Template for database services (query latency, availability)
serviceType: database

defaults:
  indicatorType: latency
  metricSource: prometheus
  budgetingPeriod: 28d
  target: 99.9
  threshold: 100ms

metrics:
  query_latency: |
    histogram_quantile(0.99,
      rate(pg_query_duration_seconds_bucket{database="$DB_NAME"}[5m])
    )
  connection_pool: |
    avg_over_time(pg_connection_pool_utilization{database="$DB_NAME"}[5m])
  replication_lag: |
    pg_replication_lag_seconds{database="$DB_NAME"}

examples:
  - name: Postgres Query Latency
    description: 99th percentile query latency
    metricQuery: histogram_quantile(0.99, rate(pg_query_duration_seconds_bucket{database="users"}[5m]))
  
  - name: Database Availability
    description: Connection success rate
    indicatorType: availability
    target: 99.95
    metricQuery: sum(rate(pg_connections_total{database="users",status="success"}[5m])) / sum(rate(pg_connections_total{database="users"}[5m]))

alerting:
  enabled: true
  notificationChannels: ["pagerduty", "slack#database-alerts", "telegram:your_chat_id"]
  burnRateThresholds:
    - rate: 10.0
      period: 30m
      severity: warning
      alertName: DatabaseWarning
    - rate: 30.0
      period: 5m
      severity: critical
      alertName: DatabaseCritical`,

		"kubernetes.yaml": `name: kubernetes
displayName: Kubernetes Infrastructure
description: Template for Kubernetes cluster components
serviceType: infrastructure

defaults:
  indicatorType: availability
  metricSource: prometheus
  budgetingPeriod: 7d
  target: 99.95
  threshold: 0.0005

metrics:
  api_server_latency: |
    histogram_quantile(0.99,
      rate(apiserver_request_duration_seconds_bucket{verb!="WATCH"}[5m])
    )
  api_server_availability: |
    sum(rate(apiserver_request_total{code=~"2.."}[5m])) / sum(rate(apiserver_request_total[5m]))
  node_ready: |
    avg_over_time(kube_node_status_condition{condition="Ready",status="true"}[5m])

examples:
  - name: Kubernetes API Server Availability
    description: Availability of Kubernetes API server
    metricQuery: sum(rate(apiserver_request_total{code=~"2.."}[5m])) / sum(rate(apiserver_request_total[5m]))
  
  - name: API Server Latency
    description: 99th percentile API server response time
    indicatorType: latency
    target: 99.9
    threshold: 200ms
    metricQuery: histogram_quantile(0.99, rate(apiserver_request_duration_seconds_bucket{verb!="WATCH"}[5m]))

alerting:
  enabled: true
  notificationChannels: ["pagerduty", "slack#k8s-alerts", "telegram:your_chat_id"]
  burnRateThresholds:
    - rate: 5.0
      period: 30m
      severity: warning
      alertName: K8sWarning
    - rate: 20.0
      period: 5m
      severity: critical
      alertName: K8sCritical`,
	}

	// Сохраняем каждый шаблон в файл
	for filename, content := range exampleTemplates {
		filePath := filepath.Join(templatesDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create template file %s: %w", filename, err)
		}
		log.Printf("Created template file: %s", filename)
	}

	return nil
}

// Обработчик списка SLO
func listSLOsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Listing SLOs")

	dataDir := "./data"

	files, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"slos": []string{}})
			return
		}
		log.Printf("Directory read error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read directory: %v", err), http.StatusInternalServerError)
		return
	}

	var slos []map[string]string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") && !strings.Contains(file.Name(), "prometheus") && !strings.Contains(file.Name(), "alertmanager") {
			slos = append(slos, map[string]string{
				"name": file.Name(),
				"path": "/data/" + file.Name(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{"slos": slos})
}

// Health check
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Основная функция
func main() {
	port := 8080

	os.MkdirAll("./data", 0755)
	os.MkdirAll("./templates", 0755) // Создаем директорию для шаблонов
	os.MkdirAll("./static", 0755)

	// Проверяем наличие Telegram переменных окружения
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	// Копируем примеры шаблонов если директория пуста
	if err := copyExampleTemplates(); err != nil {
		log.Printf("Note: %v", err)
	}

	// Статические файлы
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API endpoints
	http.HandleFunc("/", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Если запрос не API, отдаем index.html
		if !strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/data/" {
			http.ServeFile(w, r, "./static/index.html")
			return
		}
		http.NotFound(w, r)
	}))

	http.HandleFunc("/api/slo/create", loggingMiddleware(createSLOHandler))
	http.HandleFunc("/api/slos", loggingMiddleware(listSLOsHandler))
	http.HandleFunc("/api/templates", loggingMiddleware(templatesHandler))
	http.HandleFunc("/api/templates/create-defaults", loggingMiddleware(createDefaultTemplatesHandler))
	http.HandleFunc("/health", loggingMiddleware(healthHandler))

	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./data"))))

	fmt.Printf("🚀 Starting SLOzy with full alerting support on http://localhost:%d\n", port)
	fmt.Printf("📁 Data directory: ./data/\n")
	fmt.Printf("📁 Templates directory: ./templates/\n")
	fmt.Printf("📁 Static directory: ./static/\n")

	// Отображаем статус Telegram
	if telegramBotToken != "" && telegramChatID != "" {
		tokenPreview := telegramBotToken
		if len(tokenPreview) > 10 {
			tokenPreview = tokenPreview[:10] + "..."
		}
		fmt.Printf("🤖 Telegram bot configured\n")
		fmt.Printf("   • Bot token: %s\n", tokenPreview)
		fmt.Printf("   • Chat ID: %s\n", telegramChatID)
	} else {
		fmt.Printf("⚠️  Telegram bot not configured (set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID env vars)\n")
		fmt.Printf("   To enable Telegram alerts, run:\n")
		fmt.Printf("   export TELEGRAM_BOT_TOKEN='your_bot_token'\n")
		fmt.Printf("   export TELEGRAM_CHAT_ID='your_chat_id'\n")
	}

	fmt.Printf("🔔 Alerting features:\n")
	fmt.Printf("   • OpenSLO Alert Policies\n")
	fmt.Printf("   • Prometheus Alert Rules\n")
	fmt.Printf("   • AlertManager Config with Telegram support\n")
	fmt.Printf("   • Grafana Dashboards\n")
	fmt.Printf("   • README with instructions\n")
	fmt.Printf("📦 All files are now downloaded as ZIP archive\n")
	fmt.Printf("📋 Templates loaded from ./templates/\n")
	fmt.Println("\nPress Ctrl+C to stop")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal("Server failed:", err)
	}
}

// Вспомогательная функция для минимума
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
