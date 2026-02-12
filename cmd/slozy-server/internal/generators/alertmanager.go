package generators

import (
	"fmt"
	"os"
	"strings"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
	"gopkg.in/yaml.v3"
)

// AlertmanagerGenerator generates Alertmanager configuration
type AlertmanagerGenerator struct{}

// NewAlertmanagerGenerator creates a new Alertmanager generator
func NewAlertmanagerGenerator() *AlertmanagerGenerator {
	return &AlertmanagerGenerator{}
}

// Generate creates Alertmanager configuration from SLO definition
func (g *AlertmanagerGenerator) Generate(slo models.SLO) ([]byte, error) {
	if !slo.Alerting.Enabled {
		return nil, nil
	}

	config := g.buildConfig(slo)
	return yaml.Marshal(config)
}

// buildConfig builds the complete Alertmanager configuration
func (g *AlertmanagerGenerator) buildConfig(slo models.SLO) map[string]interface{} {
	// Get Telegram config from environment
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	receivers := g.buildReceivers(slo, telegramBotToken, telegramChatID)

	return map[string]interface{}{
		"global": map[string]interface{}{
			"smtp_smarthost": os.Getenv("SMTP_HOST"),
			"smtp_from":      os.Getenv("SMTP_FROM"),
		},
		"route": g.buildRoute(slo),
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
}

// buildReceivers creates notification receivers
func (g *AlertmanagerGenerator) buildReceivers(slo models.SLO, telegramBotToken, telegramChatID string) []map[string]interface{} {
	receivers := []map[string]interface{}{
		{"name": "slo-notifications"},
		{"name": "critical-alerts"},
	}

	// Parse notification channels and create configs
	telegramConfig := g.buildTelegramConfig(slo.Alerting.NotificationChannels, telegramBotToken, telegramChatID)

	if len(telegramConfig) > 0 {
		receivers[0]["telegram_configs"] = []map[string]interface{}{telegramConfig}
		receivers[1]["telegram_configs"] = []map[string]interface{}{telegramConfig}
	}

	// Add other notification types
	slackConfigs := g.buildSlackConfigs(slo.Alerting.NotificationChannels)
	emailConfigs := g.buildEmailConfigs(slo.Alerting.NotificationChannels)
	pagerdutyConfigs := g.buildPagerDutyConfigs(slo.Alerting.NotificationChannels)
	webhookConfigs := g.buildWebhookConfigs(slo.Alerting.NotificationChannels)

	if len(slackConfigs) > 0 {
		receivers[0]["slack_configs"] = slackConfigs
	}
	if len(emailConfigs) > 0 {
		receivers[0]["email_configs"] = emailConfigs
	}
	if len(pagerdutyConfigs) > 0 {
		receivers[1]["pagerduty_configs"] = pagerdutyConfigs
	}
	if len(webhookConfigs) > 0 {
		receivers[0]["webhook_configs"] = webhookConfigs
	}

	// Add default Telegram if token exists but no channels specified
	if telegramBotToken != "" && telegramChatID != "" && len(telegramConfig) == 0 {
		defaultTelegramConfig := g.buildDefaultTelegramConfig(telegramBotToken, telegramChatID)
		receivers[0]["telegram_configs"] = []map[string]interface{}{defaultTelegramConfig}
		receivers[1]["telegram_configs"] = []map[string]interface{}{defaultTelegramConfig}
	}

	return receivers
}

// buildRoute creates routing configuration
func (g *AlertmanagerGenerator) buildRoute(slo models.SLO) map[string]interface{} {
	return map[string]interface{}{
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
	}
}

// buildTelegramConfig creates Telegram notification configuration
func (g *AlertmanagerGenerator) buildTelegramConfig(channels []string, botToken, chatID string) map[string]interface{} {
	var telegramChats []string

	for _, channel := range channels {
		if strings.HasPrefix(channel, "telegram") {
			parts := strings.Split(channel, ":")
			if len(parts) > 1 && parts[1] != "" {
				telegramChats = append(telegramChats, parts[1])
			}
		}
	}

	if len(telegramChats) > 0 {
		chatID = telegramChats[0] // Use first configured chat
	}

	if botToken != "" && chatID != "" {
		return map[string]interface{}{
			"api_url":               fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
			"chat_id":               chatID,
			"message":               "🚨 *{{ .GroupLabels.severity | toUpper }} SLO Alert*\n\n*Service:* {{ .GroupLabels.service }}\n*SLO:* {{ .GroupLabels.slo }}\n*Severity:* {{ .GroupLabels.severity }}\n*Alert:* {{ .GroupLabels.alertname }}\n\n{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}\n\n{{ if (index .Alerts 0).Annotations.dashboard }}[View Dashboard]({{ (index .Alerts 0).Annotations.dashboard }}){{ end }}",
			"parse_mode":            "Markdown",
			"disable_notifications": false,
			"send_resolved":         true,
		}
	}

	return nil
}

// buildDefaultTelegramConfig creates default Telegram configuration
func (g *AlertmanagerGenerator) buildDefaultTelegramConfig(botToken, chatID string) map[string]interface{} {
	return map[string]interface{}{
		"api_url":               fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
		"chat_id":               chatID,
		"message":               "🚨 *{{ .GroupLabels.severity | toUpper }} SLO Alert*\n\n*Service:* {{ .GroupLabels.service }}\n*SLO:* {{ .GroupLabels.slo }}\n*Severity:* {{ .GroupLabels.severity }}\n\n{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}",
		"parse_mode":            "Markdown",
		"disable_notifications": false,
		"send_resolved":         true,
	}
}

// buildSlackConfigs creates Slack notification configurations
func (g *AlertmanagerGenerator) buildSlackConfigs(channels []string) []map[string]interface{} {
	var slackConfigs []map[string]interface{}

	for _, channel := range channels {
		if strings.HasPrefix(channel, "slack") {
			var slackConfig map[string]interface{}
			if strings.Contains(channel, "#") {
				parts := strings.Split(channel, "#")
				channelName := parts[1]
				slackConfig = map[string]interface{}{
					"api_url":  "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
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
		}
	}

	return slackConfigs
}

// buildEmailConfigs creates email notification configurations
func (g *AlertmanagerGenerator) buildEmailConfigs(channels []string) []map[string]interface{} {
	var emailConfigs []map[string]interface{}

	for _, channel := range channels {
		if strings.HasPrefix(channel, "email") {
			parts := strings.Split(channel, ":")
			emailTo := "sre-team@example.com"
			if len(parts) > 1 {
				emailTo = parts[1]
			}
			emailConfigs = append(emailConfigs, map[string]interface{}{
				"to":        emailTo,
				"from":      os.Getenv("SMTP_FROM"),
				"smarthost": os.Getenv("SMTP_HOST"),
				"subject":   "SLO Alert: {{ .GroupLabels.service }} {{ .GroupLabels.slo }}",
				"html":      "{{ range .Alerts }}<p><strong>{{ .Annotations.summary }}</strong><br>{{ .Annotations.description }}</p>{{ end }}",
			})
		}
	}

	return emailConfigs
}

// buildPagerDutyConfigs creates PagerDuty notification configurations
func (g *AlertmanagerGenerator) buildPagerDutyConfigs(channels []string) []map[string]interface{} {
	var pagerdutyConfigs []map[string]interface{}

	for _, channel := range channels {
		if strings.HasPrefix(channel, "pagerduty") {
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
		}
	}

	return pagerdutyConfigs
}

// buildWebhookConfigs creates webhook notification configurations
func (g *AlertmanagerGenerator) buildWebhookConfigs(channels []string) []map[string]interface{} {
	var webhookConfigs []map[string]interface{}

	for _, channel := range channels {
		if strings.HasPrefix(channel, "webhook") {
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

	return webhookConfigs
}