package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
)

func TestSafeFilename(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "slo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	handler := &SLOHandler{}

	tests := []struct {
		serviceName string
		sloName     string
		expected    string
	}{
		{"my-service", "api-latency", "my-service-api-latency"},
		{"user/profile", "api-response", "profile-api-response"},
		{"../../../etc/passwd", "malicious", "passwd-malicious"},
		{"user\x00name", "test", "username-test"},
		{"CON", "reserved", "_CON-reserved"},
		{"very-long-service-name-that-exceeds-the-maximum-allowed-length-and-should-be-truncated", "test", "very-long-service-name-that-exceeds-the-maximum-allowed-length-and-should-be-truncated-test"},
		{"normal", "unicode-test-üñïçødé", "normal-unicode-test-_____d_"},
	}

	for _, tt := range tests {
		t.Run(tt.serviceName, func(t *testing.T) {
			result := handler.safeFilename(tt.serviceName, tt.sloName)
			if result != tt.expected {
				t.Errorf("safeFilename: Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCreateSLO(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "slo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	handler := &SLOHandler{dataDir: tempDir}

	validSLO := models.SLO{
		ServiceName:  "test-service",
		SLOName:      "test-latency",
		Target:       99.9,
		MetricQuery:  "test_metric",
		IndicatorType: "latency",
		Alerting: models.AlertingConfig{
			Enabled: true,
			NotificationChannels: []string{"test"},
			BurnRateThresholds: []models.BurnRateThreshold{
				{Rate: 10.0, Period: "1h", Severity: "warning"},
			},
		},
	}

	tests := []struct {
		name           string
		slo            models.SLO
		expectedStatus int
		shouldContain  string
	}{
		{
			name:           "Valid SLO",
			slo:            validSLO,
			expectedStatus: http.StatusOK,
			shouldContain:  ".zip",
		},
		{
			name: "SLO with path traversal in name",
			slo: models.SLO{
				ServiceName: "../../../etc/passwd",
				SLOName:     "test",
				Target:      99.9,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "SLO that's too large",
			slo: models.SLO{
				ServiceName: "test",
				SLOName:     "test",
				Target:      99.9,
				MetricQuery: string(make([]byte, 10100)), // 10KB+ query
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.slo)
			req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.CreateSLO(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("CreateSLO: Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.shouldContain != "" {
				var response map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&response)
				if url, ok := response["downloadUrl"].(string); ok {
					if tt.shouldContain == ".zip" && filepath.Ext(url) != ".zip" {
						t.Errorf("CreateSLO: Expected response to contain ZIP download")
					}
				}
			}
		})
	}
}