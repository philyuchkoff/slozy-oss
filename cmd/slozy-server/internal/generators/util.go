package generators

import "os"

// grafanaBaseURL returns the Grafana base URL from environment or a default placeholder.
func grafanaBaseURL() string {
	if url := os.Getenv("GRAFANA_URL"); url != "" {
		return url
	}
	return "https://grafana.example.com"
}
