package generators

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
)

// GrafanaGenerator generates Grafana dashboard configurations
type GrafanaGenerator struct {
	openSLOGen *OpenSLOGenerator
}

// NewGrafanaGenerator creates a new Grafana generator
func NewGrafanaGenerator() *GrafanaGenerator {
	return &GrafanaGenerator{
		openSLOGen: NewOpenSLOGenerator(),
	}
}

// Generate creates a Grafana dashboard from SLO definition
func (g *GrafanaGenerator) Generate(slo models.SLO) ([]byte, error) {
	dashboard := map[string]interface{}{
		"dashboard": g.buildMainDashboard(slo),
		"folderId":  0,
		"overwrite": true,
	}

	return json.MarshalIndent(dashboard, "", "  ")
}

// generateMultiBurnRateDashboard creates a dashboard with multiple burn rate thresholds
func (g *GrafanaGenerator) GenerateMultiBurnRate(slo models.SLO) ([]byte, error) {
	dashboard := map[string]interface{}{
		"dashboard": g.buildMultiBurnRateDashboard(slo),
	}

	return json.MarshalIndent(dashboard, "", "  ")
}

// buildMainDashboard builds the main SLO dashboard
func (g *GrafanaGenerator) buildMainDashboard(slo models.SLO) map[string]interface{} {
	return map[string]interface{}{
		"title":    fmt.Sprintf("%s - %s SLO", slo.ServiceName, slo.SLOName),
		"tags":     []string{"slo", slo.ServiceName, slo.Team, slo.IndicatorType},
		"timezone": "browser",
		"panels": []map[string]interface{}{
			g.buildBurnRatePanel(slo),
			g.buildErrorBudgetPanel(slo),
			g.buildAlertHistoryPanel(),
		},
		"templating": g.buildTemplating(slo),
		"refresh":       "30s",
		"schemaVersion": 27,
		"version":       1,
		"uid":           fmt.Sprintf("%s-%s-slo", strings.Replace(slo.ServiceName, " ", "-", -1), strings.Replace(slo.SLOName, " ", "-", -1)),
	}
}

// buildMultiBurnRateDashboard builds the multi burn rate dashboard
func (g *GrafanaGenerator) buildMultiBurnRateDashboard(slo models.SLO) map[string]interface{}{
	dashboard := map[string]interface{}{
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
	}

	panels := dashboard["panels"].([]map[string]interface{})

	for i, br := range slo.Alerting.BurnRateThresholds {
		panel := g.buildBurnRateThresholdPanel(slo, br, i)
		panels = append(panels, panel)
	}

	return dashboard
}

// buildBurnRatePanel creates the burn rate panel
func (g *GrafanaGenerator) buildBurnRatePanel(slo models.SLO) map[string]interface{} {
	return map[string]interface{}{
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
		"alert": g.buildBurnRateAlert(),
	}
}

// buildErrorBudgetPanel creates the error budget panel
func (g *GrafanaGenerator) buildErrorBudgetPanel(slo models.SLO) map[string]interface{} {
	return map[string]interface{}{
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
		"options": g.buildErrorBudgetOptions(),
	}
}

// buildAlertHistoryPanel creates the alert history panel
func (g *GrafanaGenerator) buildAlertHistoryPanel() map[string]interface{} {
	return map[string]interface{}{
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
	}
}

// buildBurnRateThresholdPanel creates a panel for a specific burn rate threshold
func (g *GrafanaGenerator) buildBurnRateThresholdPanel(slo models.SLO, br models.BurnRateThreshold, index int) map[string]interface{} {
	return map[string]interface{}{
		"title": fmt.Sprintf("%s Burn Rate (%s)", strings.Title(br.Severity), br.Period),
		"type":  "graph",
		"gridPos": map[string]interface{}{
			"h": 8, "w": 12, "x": (index % 2) * 12, "y": int(math.Floor(float64(index)/2)) * 8,
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
}

// buildTemplating creates templating variables
func (g *GrafanaGenerator) buildTemplating(slo models.SLO) map[string]interface{} {
	return map[string]interface{}{
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
	}
}

// buildBurnRateAlert creates burn rate alert configuration
func (g *GrafanaGenerator) buildBurnRateAlert() map[string]interface{} {
	return map[string]interface{}{
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
	}
}

// buildErrorBudgetOptions creates error budget panel options
func (g *GrafanaGenerator) buildErrorBudgetOptions() map[string]interface{} {
	return map[string]interface{}{
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
	}
}