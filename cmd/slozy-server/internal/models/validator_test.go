package models

import (
	"testing"
)

func TestValidator(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		slo       SLO
		expectErr bool
		errFields map[string]bool
	}{
		{
			name: "Valid SLO",
			slo: SLO{
				ServiceName:  "test-service",
				SLOName:      "test-latency",
				Target:       99.9,
				MetricQuery:  "test_metric",
				IndicatorType: "latency",
			},
			expectErr: false,
		},
		{
			name: "Missing required fields",
			slo: SLO{
				Target: 99.9,
			},
			expectErr: true,
			errFields: map[string]bool{
				"Service name": true,
				"SLO name":     true,
				"Metric query": true,
			},
		},
		{
			name: "Invalid target",
			slo: SLO{
				ServiceName:  "test-service",
				SLOName:      "test-latency",
				Target:       85.0, // Below minimum
				MetricQuery:  "test_metric",
				IndicatorType: "latency",
			},
			expectErr: true,
			errFields: map[string]bool{
				"target": true,
			},
		},
		{
			name: "Names too long",
			slo: SLO{
				ServiceName:  string(make([]byte, 101)), // 101 chars
				SLOName:      "test-latency",
				Target:       99.9,
				MetricQuery:  "test_metric",
				IndicatorType: "latency",
			},
			expectErr: true,
			errFields: map[string]bool{
				"Service name": true,
			},
		},
		{
			name: "Alerting with no channels",
			slo: SLO{
				ServiceName:  "test-service",
				SLOName:      "test-latency",
				Target:       99.9,
				MetricQuery:  "test_metric",
				IndicatorType: "latency",
				Alerting: AlertingConfig{
					Enabled: true,
					// No notification channels
				},
			},
			expectErr: true,
			errFields: map[string]bool{
				"alerting.notificationChannels": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.Validate(tt.slo)
			hasErr := len(errors) > 0

			if hasErr != tt.expectErr {
				t.Errorf("Validator: Expected error: %v, got: %v", tt.expectErr, hasErr)
			}

			if tt.errFields != nil {
				errFields := make(map[string]bool)
				for _, err := range errors {
					errFields[err.Field] = true
				}

				for expectedField := range tt.errFields {
					if !errFields[expectedField] {
						t.Errorf("Validator: Expected error field %q, but not found", expectedField)
					}
				}
			}
		})
	}
}