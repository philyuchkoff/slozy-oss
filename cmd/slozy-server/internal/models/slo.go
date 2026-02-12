package models

import "time"

// SLO represents a Service Level Objective
type SLO struct {
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

// AlertingConfig contains alerting configuration for an SLO
type AlertingConfig struct {
	Enabled              bool                `json:"enabled"`
	NotificationChannels []string            `json:"notificationChannels"`
	BurnRateThresholds   []BurnRateThreshold `json:"burnRateThresholds"`
}

// BurnRateThreshold defines a burn rate alert threshold
type BurnRateThreshold struct {
	Rate      float64 `json:"rate"`      // 10.0 (10x быстрее)
	Period    string  `json:"period"`    // "1h", "6h"
	Severity  string  `json:"severity"`  // "warning", "critical"
	AlertName string  `json:"alertName"` // "HighBurnRate"
}

// TemplateConfig represents a template configuration from YAML
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

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Config holds server configuration
type Config struct {
	Port                int
	DataDirectory       string
	TemplatesDirectory  string
	StaticDirectory     string
	TelegramBotToken    string
	TelegramChatID      string
	SMTPHost            string
	SMTPFrom            string
	GrafanaURL          string
}

// HealthResponse is the response for health checks
type HealthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Errors  []string    `json:"errors,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}