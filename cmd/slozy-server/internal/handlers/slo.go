package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/generators"
	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/middleware"
)

// SLOHandler handles SLO creation and related operations
type SLOHandler struct {
	validator         *models.Validator
	openSLOGen        *generators.OpenSLOGenerator
	prometheusGen     *generators.PrometheusGenerator
	alertmanagerGen   *generators.AlertmanagerGenerator
	grafanaGen        *generators.GrafanaGenerator
	readmeGen         *generators.ReadmeGenerator
	tempDir           string
	dataDir           string
}

// NewSLOHandler creates a new SLO handler
func NewSLOHandler(dataDir string) *SLOHandler {
	return &SLOHandler{
		validator:       models.NewValidator(),
		openSLOGen:      generators.NewOpenSLOGenerator(),
		prometheusGen:   generators.NewPrometheusGenerator(),
		alertmanagerGen: generators.NewAlertmanagerGenerator(),
		grafanaGen:      generators.NewGrafanaGenerator(),
		readmeGen:       generators.NewReadmeGenerator(),
		dataDir:        dataDir,
		tempDir:        dataDir + "/tmp",
	}
}

// CreateSLO handles SLO creation requests
func (h *SLOHandler) CreateSLO(w http.ResponseWriter, r *http.Request) {
	log.Println("Received SLO creation request")

	if r.Method != http.MethodPost {
		log.Printf("Wrong method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var slo models.SLO
	if err := json.NewDecoder(r.Body).Decode(&slo); err != nil {
		log.Printf("JSON decode error: %v", err)
		respondWithError(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Parsed SLO data: Service: %s, SLO: %s", slo.ServiceName, slo.SLOName)

	if validationErrors := h.validator.ValidateLegacy(slo); len(validationErrors) > 0 {
		log.Printf("Validation failed: %v", validationErrors)
		respondWithValidationErrors(w, validationErrors)
		return
	}

	// Apply defaults
	h.applyDefaults(&slo)

	// Create temporary directory for files
	tempDir, err := os.MkdirTemp("", "slozy-*")
	if err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		respondWithError(w, fmt.Sprintf("Failed to create temp directory: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	// Generate all files
	zipData, filename, err := h.generateAndZip(tempDir, slo)
	if err != nil {
		log.Printf("Failed to generate files: %v", err)
		respondWithError(w, fmt.Sprintf("Failed to generate files: %v", err), http.StatusInternalServerError)
		return
	}

	// Send ZIP archive
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

	if _, err := w.Write(zipData); err != nil {
		log.Printf("Failed to send ZIP archive: %v", err)
	}

	log.Printf("ZIP archive sent successfully: %s (%d bytes)", filename, len(zipData))
}

// generateAndZip generates all files and creates ZIP archive
func (h *SLOHandler) generateAndZip(tempDir string, slo models.SLO) ([]byte, string, error) {
	baseFilename := h.safeFilename(slo.ServiceName, slo.SLOName)

	// Generate main SLO YAML
	yamlContent, err := h.openSLOGen.Generate(slo)
	if err != nil {
		return nil, "", fmt.Errorf("YAML generation error: %w", err)
	}

	mainFilename := baseFilename + ".yaml"
	if err := os.WriteFile(filepath.Join(tempDir, mainFilename), yamlContent, 0644); err != nil {
		return nil, "", fmt.Errorf("file write error: %w", err)
	}

	allFiles := map[string][]byte{mainFilename: yamlContent}
	additionalFiles := make(map[string]string)

	// Generate alerting files if enabled
	if slo.Alerting.Enabled {
		if err := h.generateAlertingFiles(tempDir, slo, baseFilename, allFiles, additionalFiles); err != nil {
			return nil, "", err
		}
	}

	// Generate README
	if slo.Alerting.Enabled {
		readmeContent := h.readmeGen.Generate(slo, mainFilename, additionalFiles)
		readmeFilename := baseFilename + "-README.md"
		if err := os.WriteFile(filepath.Join(tempDir, readmeFilename), []byte(readmeContent), 0644); err == nil {
			allFiles[readmeFilename] = []byte(readmeContent)
		}
	}

	// Create ZIP archive
	zipData, err := h.createZipArchive(allFiles)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create ZIP archive: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.zip", baseFilename, timestamp)

	return zipData, filename, nil
}

// generateAlertingFiles generates all alerting related files
func (h *SLOHandler) generateAlertingFiles(tempDir string, slo models.SLO, baseFilename string, allFiles map[string][]byte, additionalFiles map[string]string) error {
	// Prometheus Alert Rules
	if prometheusRules, err := h.prometheusGen.Generate(slo); err == nil && prometheusRules != nil {
		promFilename := baseFilename + "-prometheus-alerts.yaml"
		if err := os.WriteFile(filepath.Join(tempDir, promFilename), prometheusRules, 0644); err == nil {
			allFiles[promFilename] = prometheusRules
			additionalFiles["prometheusAlerts"] = promFilename
		}
	}

	// Alertmanager Config
	if alertManagerConfig, err := h.alertmanagerGen.Generate(slo); err == nil && alertManagerConfig != nil {
		amFilename := baseFilename + "-alertmanager-config.yaml"
		if err := os.WriteFile(filepath.Join(tempDir, amFilename), alertManagerConfig, 0644); err == nil {
			allFiles[amFilename] = alertManagerConfig
			additionalFiles["alertManagerConfig"] = amFilename
		}
	}

	// Grafana Dashboards
	if grafanaDashboard, err := h.grafanaGen.Generate(slo); err == nil && grafanaDashboard != nil {
		grafanaFilename := baseFilename + "-grafana-dashboard.json"
		if err := os.WriteFile(filepath.Join(tempDir, grafanaFilename), grafanaDashboard, 0644); err == nil {
			allFiles[grafanaFilename] = grafanaDashboard
			additionalFiles["grafanaDashboard"] = grafanaFilename
		}
	}

	if multiBurnRateDashboard, err := h.grafanaGen.GenerateMultiBurnRate(slo); err == nil && multiBurnRateDashboard != nil {
		multiFilename := baseFilename + "-multi-burnrate-dashboard.json"
		if err := os.WriteFile(filepath.Join(tempDir, multiFilename), multiBurnRateDashboard, 0644); err == nil {
			allFiles[multiFilename] = multiBurnRateDashboard
		}
	}

	return nil
}

// applyDefaults applies default values to SLO
func (h *SLOHandler) applyDefaults(slo *models.SLO) {
	if slo.BudgetingPeriod == "" {
		slo.BudgetingPeriod = "28d"
	}
	if slo.IndicatorType == "" {
		slo.IndicatorType = "latency"
	}
	if slo.MetricSource == "" {
		slo.MetricSource = "prometheus"
	}
	if slo.Threshold == "" {
		slo.Threshold = "200ms"
	}
	if slo.Team == "" {
		slo.Team = "unknown"
	}
}

// safeFilename creates a safe filename from service and SLO names
func (h *SLOHandler) safeFilename(serviceName, sloName string) string {
	safeFunc := func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '_'
	}

	safeService := strings.Map(safeFunc, serviceName)
	safeSLO := strings.Map(safeFunc, sloName)
	return fmt.Sprintf("%s-%s", safeService, safeSLO)
}

// createZipArchive creates ZIP archive from file map
func (h *SLOHandler) createZipArchive(files map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create file in zip: %w", err)
		}
		if _, err := writer.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write file to zip: %w", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// respondWithError sends error response
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// respondWithValidationErrors sends validation error response
func respondWithValidationErrors(w http.ResponseWriter, errors []string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"errors":  errors,
	})
}

// ListSLOs handles listing existing SLOs
func (h *SLOHandler) ListSLOs(w http.ResponseWriter, r *http.Request) {
	log.Println("Listing SLOs")

	files, err := os.ReadDir(h.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			json.NewEncoder(w).Encode(map[string]interface{}{"slos": []string{}})
			return
		}
		log.Printf("Directory read error: %v", err)
		respondWithError(w, fmt.Sprintf("Failed to read directory: %v", err), http.StatusInternalServerError)
		return
	}

	var slos []map[string]string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") &&
			!strings.Contains(file.Name(), "prometheus") && !strings.Contains(file.Name(), "alertmanager") {
			slos = append(slos, map[string]string{
				"name": file.Name(),
				"path": "/data/" + file.Name(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{"slos": slos})
}

// Wrap creates middleware wrapper
func (h *SLOHandler) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return middleware.LoggingMiddleware(next)
}