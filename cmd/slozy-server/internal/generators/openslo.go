package generators

import (
	"fmt"
	"strings"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
	"gopkg.in/yaml.v3"
)

// OpenSLOGenerator generates OpenSLO compliant YAML files
type OpenSLOGenerator struct{}

// NewOpenSLOGenerator creates a new OpenSLO generator
func NewOpenSLOGenerator() *OpenSLOGenerator {
	return &OpenSLOGenerator{}
}

// Generate creates an OpenSLO YAML document from the SLO definition
func (g *OpenSLOGenerator) Generate(slo models.SLO) ([]byte, error) {
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
	spec["indicator"] = g.buildIndicator(slo)

	// Add alert policies if enabled
	if slo.Alerting.Enabled && len(slo.Alerting.BurnRateThresholds) > 0 {
		spec["alertPolicies"] = g.buildAlertPolicies(slo)
	}

	return yaml.Marshal(openslo)
}

// buildIndicator creates the indicator configuration based on indicator type
func (g *OpenSLOGenerator) buildIndicator(slo models.SLO) map[string]interface{} {
	switch strings.ToLower(slo.IndicatorType) {
	case "latency":
		return map[string]interface{}{
			"thresholdMetric": map[string]interface{}{
				"threshold": map[string]interface{}{
					"metricSourceRef": slo.MetricSource,
					"query":           slo.MetricQuery,
				},
			},
		}
	case "availability":
		return map[string]interface{}{
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
		return map[string]interface{}{
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
}

// buildAlertPolicies creates alert policies for the SLO
func (g *OpenSLOGenerator) buildAlertPolicies(slo models.SLO) []map[string]interface{} {
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

	return alertPolicies
}

// formatTarget formats target for beautiful output
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