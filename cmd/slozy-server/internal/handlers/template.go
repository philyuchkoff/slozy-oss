package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/middleware"
	"gopkg.in/yaml.v3"
)

// TemplateHandler handles template operations
type TemplateHandler struct {
	templatesDir string
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(templatesDir string) *TemplateHandler {
	return &TemplateHandler{
		templatesDir: templatesDir,
	}
}

// GetTemplates returns all available templates
func (h *TemplateHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.loadTemplates()
	if err != nil {
		log.Printf("Warning: Failed to load templates from files, using defaults: %v", err)
		templates = h.getDefaultTemplates()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{"templates": templates})
}

// CreateDefaultTemplates creates default template files
func (h *TemplateHandler) CreateDefaultTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.copyExampleTemplates(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Default templates created successfully",
	})
}

// loadTemplates loads templates from files
func (h *TemplateHandler) loadTemplates() ([]map[string]interface{}, error) {
	if err := os.MkdirAll(h.templatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	files, err := os.ReadDir(h.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []map[string]interface{}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			template, err := h.loadTemplateFile(filepath.Join(h.templatesDir, file.Name()))
			if err != nil {
				log.Printf("Warning: Failed to load template file %s: %v", file.Name(), err)
				continue
			}
			if template != nil {
				templates = append(templates, template)
			}
		}
	}

	if len(templates) == 0 {
		return h.getDefaultTemplates(), nil
	}

	return templates, nil
}

// loadTemplateFile loads a single template file
func (h *TemplateHandler) loadTemplateFile(filePath string) (map[string]interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var template models.TemplateConfig
	if err := yaml.Unmarshal(content, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template file: %w", err)
	}

	// Convert to frontend format
	frontendTemplate := h.convertToFrontendTemplate(template)
	return frontendTemplate, nil
}

// convertToFrontendTemplate converts internal template to frontend format
func (h *TemplateHandler) convertToFrontendTemplate(template models.TemplateConfig) map[string]interface{} {
	frontendTemplate := map[string]interface{}{
		"name":        template.Name,
		"displayName": template.DisplayName,
		"description": template.Description,
		"serviceType": template.ServiceType,
		"defaults":    template.Defaults,
	}

	// Add alerting if enabled
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

	// Add examples to defaults
	if len(template.Examples) > 0 {
		if frontendTemplate["defaults"] == nil {
			frontendTemplate["defaults"] = make(map[string]interface{})
		}
		defaults := frontendTemplate["defaults"].(map[string]interface{})

		example := template.Examples[0]
		defaults["metricQuery"] = example.MetricQuery

		if example.IndicatorType != "" {
			defaults["indicatorType"] = example.IndicatorType
		}
		if example.Target > 0 {
			defaults["target"] = example.Target
		}
		if example.Threshold != "" {
			defaults["threshold"] = example.Threshold
		}
	}

	// Add serviceType to defaults
	if frontendTemplate["defaults"] == nil {
		frontendTemplate["defaults"] = make(map[string]interface{})
	}
	frontendTemplate["defaults"].(map[string]interface{})["serviceType"] = template.ServiceType

	return frontendTemplate
}

// getDefaultTemplates returns default template definitions
func (h *TemplateHandler) getDefaultTemplates() []map[string]interface{} {
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

// copyExampleTemplates creates example template files
func (h *TemplateHandler) copyExampleTemplates() error {
	// Check if directory is not empty
	files, err := os.ReadDir(h.templatesDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	if len(files) > 0 {
		return fmt.Errorf("templates directory is not empty")
	}

	// Create example templates
	exampleTemplates := h.getExampleTemplates()

	for filename, content := range exampleTemplates {
		filePath := filepath.Join(h.templatesDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create template file %s: %w", filename, err)
		}
		log.Printf("Created template file: %s", filename)
	}

	return nil
}

// getExampleTemplates returns example template contents
func (h *TemplateHandler) getExampleTemplates() map[string]string {
	return map[string]string{
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
	}
}

// Wrap creates middleware wrapper
func (h *TemplateHandler) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return middleware.LoggingMiddleware(next)
}