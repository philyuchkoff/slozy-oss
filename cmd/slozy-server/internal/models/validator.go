package models

import (
	"fmt"
	"strings"
)

// Validator handles SLO validation
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates an SLO definition
func (v *Validator) Validate(slo SLO) []ValidationError {
	var errors []ValidationError

	if err := v.validateRequiredField(slo.ServiceName, "Service name"); err != nil {
		errors = append(errors, *err)
	}

	if err := v.validateRequiredField(slo.SLOName, "SLO name"); err != nil {
		errors = append(errors, *err)
	}

	if err := v.validateTarget(slo.Target); err != nil {
		errors = append(errors, *err)
	}

	if err := v.validateRequiredField(slo.MetricQuery, "Metric query"); err != nil {
		errors = append(errors, *err)
	}

	// Validate metric query length
	if len(slo.MetricQuery) > 10000 {
		errors = append(errors, ValidationError{
			Field:   "Metric query",
			Message: "Metric query must be less than 10000 characters",
		})
	}

	// Validate Prometheus query for injection prevention
	if err := v.validatePrometheusQuery(slo.MetricQuery); err != nil {
		errors = append(errors, *err)
	}

	if err := v.validateIndicatorType(slo.IndicatorType); err != nil {
		errors = append(errors, *err)
	}

	// Validate Team length if provided
	if slo.Team != "" {
		if err := v.validateFieldLength(slo.Team, "Team"); err != nil {
			errors = append(errors, *err)
		}
	}

	// Validate Threshold length if provided
	if slo.Threshold != "" {
		if err := v.validateFieldLength(slo.Threshold, "Threshold"); err != nil {
			errors = append(errors, *err)
		}
	}

	if err := v.validateAlerting(slo.Alerting); err != nil {
		errors = append(errors, *err)
	}

	return errors
}

// validatePrometheusQuery validates Prometheus query for injection prevention
func (v *Validator) validatePrometheusQuery(query string) *ValidationError {
	// Define allowed characters in PromQL
	// Letters, numbers, underscores, spaces, and commonly used PromQL symbols
	allowedPattern := []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'_', '-', '+', '-', '*', '/', '=', '!', '<', '>', '(', ')',
		'[', ']', '{', '}', ':', '.', ',', ' ', '\t', '\n',
		'"', '\'', '$', '|', '^',
	}

	// Create a set of allowed characters for fast lookup
	allowed := make(map[byte]bool)
	for _, c := range allowedPattern {
		allowed[c] = true
	}

	// Check each character in query
	for i, char := range []byte(query) {
		if !allowed[char] {
			return &ValidationError{
				Field:   "Metric query",
				Message: fmt.Sprintf("Invalid character at position %d: '%c'. Only standard PromQL characters are allowed", i, char),
			}
		}
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"exec", "system", "shell", "cmd", "powershell", "bash", "sh",
		"eval", "import", "require", "function", "=>", "{", "}",
		"`", ";", "&&", "||", ">", ">>", "<", "<<",
	}

	queryLower := strings.ToLower(query)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(queryLower, pattern) {
			return &ValidationError{
				Field:   "Metric query",
				Message: fmt.Sprintf("Dangerous pattern detected: '%s'. This pattern is not allowed in Prometheus queries", pattern),
			}
		}
	}

	// Additional check: ensure no command substitution
	if strings.Contains(query, "$(") || strings.Contains(query, "`") {
		return &ValidationError{
			Field:   "Metric query",
			Message: "Command substitution is not allowed in Prometheus queries",
		}
	}

	return nil
}

// ValidateLegacy validates SLO and returns string messages (for backward compatibility)
func (v *Validator) ValidateLegacy(slo SLO) []string {
	errors := v.Validate(slo)
	result := make([]string, len(errors))
	for i, err := range errors {
		result[i] = err.Message
	}
	return result
}

// validateRequiredField validates a required string field
func (v *Validator) validateRequiredField(value, fieldName string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " is required",
		}
	}

	// Add length validation
	if err := v.validateFieldLength(value, fieldName); err != nil {
		return err
	}

	return nil
}

// validateFieldLength validates string field length
func (v *Validator) validateFieldLength(value, fieldName string) *ValidationError {
	// Define maximum lengths for different fields
	maxLengths := map[string]int{
		"Service name": 100,
		"SLO name":     100,
		"Team":         100,
		"Threshold":    50,
	}

	// Field-specific validation
	if maxLength, exists := maxLengths[fieldName]; exists {
		if len(value) > maxLength {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be less than %d characters", fieldName, maxLength),
			}
		}
	}

	// General validation - no field should exceed 1000 characters
	if len(value) > 1000 {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s is too long (max 1000 characters)", fieldName),
		}
	}

	return nil
}

// validateTarget validates the SLO target value
func (v *Validator) validateTarget(target float64) *ValidationError {
	if target < 90 || target > 100 {
		return &ValidationError{
			Field:   "target",
			Message: "Target must be between 90 and 100",
		}
	}
	return nil
}

// validateIndicatorType validates the indicator type
func (v *Validator) validateIndicatorType(indicatorType string) *ValidationError {
	if indicatorType != "" {
		validTypes := map[string]bool{
			"latency":      true,
			"availability":  true,
			"throughput":    true,
			"error_rate":    true,
		}

		if !validTypes[strings.ToLower(indicatorType)] {
			return &ValidationError{
				Field:   "indicatorType",
				Message: "Indicator type must be one of: latency, availability, throughput, error_rate",
			}
		}
	}
	return nil
}

// validateAlerting validates alerting configuration
func (v *Validator) validateAlerting(alerting AlertingConfig) *ValidationError {
	if alerting.Enabled {
		if len(alerting.NotificationChannels) == 0 {
			return &ValidationError{
				Field:   "alerting.notificationChannels",
				Message: "At least one notification channel must be configured when alerting is enabled",
			}
		}

		if len(alerting.BurnRateThresholds) == 0 {
			return &ValidationError{
				Field:   "alerting.burnRateThresholds",
				Message: "At least one burn rate threshold must be configured when alerting is enabled",
			}
		}

		// Validate burn rate thresholds
		for i, br := range alerting.BurnRateThresholds {
			if br.Rate <= 0 {
				return &ValidationError{
					Field:   fmt.Sprintf("alerting.burnRateThresholds[%d].rate", i),
					Message: "Burn rate must be greater than 0",
				}
			}

			if br.Severity != "" {
				validSeverities := map[string]bool{
					"warning":  true,
					"critical":  true,
					"info":      true,
				}

				if !validSeverities[br.Severity] {
					return &ValidationError{
						Field:   fmt.Sprintf("alerting.burnRateThresholds[%d].severity", i),
						Message: "Severity must be one of: warning, critical, info",
					}
				}
			}
		}
	}
	return nil
}