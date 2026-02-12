package models

import "strings"

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

	if err := v.validateIndicatorType(slo.IndicatorType); err != nil {
		errors = append(errors, *err)
	}

	if err := v.validateAlerting(slo.Alerting); err != nil {
		errors = append(errors, *err)
	}

	return errors
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
					Field:   "alerting.burnRateThresholds[" + string(i) + "].rate",
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
						Field:   "alerting.burnRateThresholds[" + string(i) + "].severity",
						Message: "Severity must be one of: warning, critical, info",
					}
				}
			}
		}
	}
	return nil
}