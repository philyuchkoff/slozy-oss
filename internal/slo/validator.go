// валидация - чего они там вводят
package slo

import (
	"fmt"
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

type Validator struct {
	allowedIndicators    map[string]bool
	allowedMetricSources map[string]bool
	allowedPeriods       map[string]bool
}

func NewValidator() *Validator {
	return &Validator{
		allowedIndicators: map[string]bool{
			"latency":      true,
			"availability": true,
			"throughput":   true,
			"custom":       true,
		},
		allowedMetricSources: map[string]bool{
			"prometheus":  true,
			"datadog":     true,
			"cloudwatch":  true,
			"stackdriver": true,
			"newrelic":    true,
		},
		allowedPeriods: map[string]bool{
			"28d": true,
			"30d": true,
			"7d":  true,
			"1d":  true,
		},
	}
}

func (v *Validator) Validate(slo ServiceSLO) []ValidationError {
	var errors []ValidationError

	// Required fields
	if strings.TrimSpace(slo.ServiceName) == "" {
		errors = append(errors, ValidationError{
			Field:   "serviceName",
			Message: "Service name is required",
		})
	} else if !isValidName(slo.ServiceName) {
		errors = append(errors, ValidationError{
			Field:   "serviceName",
			Message: "Service name can only contain letters, numbers, and hyphens",
		})
	}

	if strings.TrimSpace(slo.SLOName) == "" {
		errors = append(errors, ValidationError{
			Field:   "sloName",
			Message: "SLO name is required",
		})
	}

	// Indicator validation
	if !v.allowedIndicators[strings.ToLower(slo.IndicatorType)] {
		errors = append(errors, ValidationError{
			Field:   "indicatorType",
			Message: fmt.Sprintf("Invalid indicator type. Allowed: %v", v.allowedIndicators),
		})
	}

	// Metric source validation
	if !v.allowedMetricSources[strings.ToLower(slo.MetricSource)] {
		errors = append(errors, ValidationError{
			Field:   "metricSource",
			Message: fmt.Sprintf("Invalid metric source. Allowed: %v", v.allowedMetricSources),
		})
	}

	// Metric query validation
	if strings.TrimSpace(slo.MetricQuery) == "" {
		errors = append(errors, ValidationError{
			Field:   "metricQuery",
			Message: "Metric query is required",
		})
	}

	// Target validation
	if slo.Target < 90 || slo.Target > 100 {
		errors = append(errors, ValidationError{
			Field:   "target",
			Message: "Target must be between 90 and 100",
		})
	}

	// Budgeting period validation
	if !v.allowedPeriods[slo.BudgetingPeriod] {
		errors = append(errors, ValidationError{
			Field:   "budgetingPeriod",
			Message: fmt.Sprintf("Invalid budgeting period. Allowed: %v", v.allowedPeriods),
		})
	}

	// Threshold validation based on indicator type
	if slo.IndicatorType == "latency" {
		if !isValidDuration(slo.Threshold) {
			errors = append(errors, ValidationError{
				Field:   "threshold",
				Message: "For latency indicators, threshold must be a duration (e.g., 200ms, 1s, 500ms)",
			})
		}
	}

	return errors
}

func isValidName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9-]+$`, name)
	return matched
}

func isValidDuration(duration string) bool {
	matched, _ := regexp.MatchString(`^\d+(\.\d+)?(ms|s|m|h)$`, duration)
	return matched
}
