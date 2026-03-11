package generators

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
	"gopkg.in/yaml.v3"
)

// PrometheusGenerator generates Prometheus alert rules
type PrometheusGenerator struct {
	openSLOGen *OpenSLOGenerator
}

// NewPrometheusGenerator creates a new Prometheus generator
func NewPrometheusGenerator() *PrometheusGenerator {
	return &PrometheusGenerator{
		openSLOGen: NewOpenSLOGenerator(),
	}
}

// Generate creates Prometheus alert rules from SLO definition
func (g *PrometheusGenerator) Generate(slo models.SLO) ([]byte, error) {
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
		alertRule := g.buildAlertRule(slo, br, targetValue)
		group["rules"] = append(group["rules"].([]map[string]interface{}), alertRule)

		// Add critical alert for warning thresholds with burn rate < 100
		if br.Severity == "warning" && br.Rate < 100 {
			criticalAlertRule := g.buildCriticalAlertRule(slo, br, targetValue)
			group["rules"] = append(group["rules"].([]map[string]interface{}), criticalAlertRule)
		}
	}

	return yaml.Marshal(rules)
}

// buildAlertRule creates a single alert rule
func (g *PrometheusGenerator) buildAlertRule(slo models.SLO, br models.BurnRateThreshold, targetValue string) map[string]interface{} {
	expr := g.buildExpression(slo, targetValue, br.Rate)

	alertName := br.AlertName
	if alertName == "" {
		alertName = fmt.Sprintf("%s_%s_BurnRateHigh",
			strings.Replace(slo.ServiceName, "-", "_", -1),
			strings.Replace(slo.SLOName, "-", "_", -1))
	}

	// Create labels with notification channel information
	labels := map[string]string{
		"severity":            br.Severity,
		"service":             slo.ServiceName,
		"slo":                 slo.SLOName,
		"slo_indicator":       slo.IndicatorType,
		"slo_target":          fmt.Sprintf("%.1f", slo.Target),
		"team":                slo.Team,
		"burn_rate_threshold": fmt.Sprintf("%.1fx", br.Rate),
	}

	// Add notification channels if present
	if len(slo.Alerting.NotificationChannels) > 0 {
		labels["notification_channels"] = strings.Join(slo.Alerting.NotificationChannels, ",")
	}

	return map[string]interface{}{
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
}

// buildCriticalAlertRule creates a critical alert rule for warning thresholds
func (g *PrometheusGenerator) buildCriticalAlertRule(slo models.SLO, br models.BurnRateThreshold, targetValue string) map[string]interface{} {
	expr := g.buildExpression(slo, targetValue, br.Rate)
	baseExpr := strings.TrimSuffix(expr, fmt.Sprintf("> %.1f", br.Rate))

	return map[string]interface{}{
		"alert": br.AlertName + "_Critical",
		"expr":  fmt.Sprintf("%s > %.1f", baseExpr, br.Rate*5),
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
}

// buildExpression creates Prometheus expression based on indicator type
func (g *PrometheusGenerator) buildExpression(slo models.SLO, targetValue string, rate float64) string {
	// Sanitize the metric query as a defense-in-depth measure
	sanitizedQuery := g.sanitizeQuery(slo.MetricQuery)

	switch slo.IndicatorType {
	case "latency":
		return fmt.Sprintf(`(
                (1 - (clamp_min(%s, 0) / %s))
                / (1 - %s)
            ) > %.1f`,
			sanitizedQuery,
			slo.Threshold,
			targetValue,
			rate)
	default:
		return fmt.Sprintf(`(
                (1 - (clamp_min(%s, 0)))
                / (1 - %s)
            ) > %.1f`,
			sanitizedQuery,
			targetValue,
			rate)
	}
}

// sanitizeQuery sanitizes a Prometheus metric query for safe inclusion in expressions
func (g *PrometheusGenerator) sanitizeQuery(query string) string {
	// Remove any potentially dangerous characters that weren't caught by validation
	// This is a defense-in-depth measure
	dangerousChars := regexp.MustCompile(`[;&|` + "`" + `]`)
	sanitized := dangerousChars.ReplaceAllString(query, "")

	// Ensure query is properly quoted if it contains spaces
	if strings.Contains(sanitized, " ") && !strings.HasPrefix(sanitized, `"`) {
		sanitized = `"` + sanitized + `"`
	}

	return sanitized
}