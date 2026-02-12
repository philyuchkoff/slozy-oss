package generators

import (
	"fmt"
	"strings"
	"time"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
)

// ReadmeGenerator generates README files for SLO configurations
type ReadmeGenerator struct{}

// NewReadmeGenerator creates a new README generator
func NewReadmeGenerator() *ReadmeGenerator {
	return &ReadmeGenerator{}
}

// Generate creates README content for an SLO
func (g *ReadmeGenerator) Generate(slo models.SLO, mainFilename string, additionalFiles map[string]string) string {
	var readme strings.Builder

	g.writeHeader(&readme, slo)
	g.writeServiceInfo(&readme, slo)
	g.writeFilesSection(&readme, mainFilename, additionalFiles)

	if slo.Alerting.Enabled && len(slo.Alerting.BurnRateThresholds) > 0 {
		g.writeAlertingSection(&readme, slo)
	}

	g.writeErrorBudgetSection(&readme, slo)
	g.writeLinks(&readme)

	return readme.String()
}

// writeHeader writes README header
func (g *ReadmeGenerator) writeHeader(readme *strings.Builder, slo models.SLO) {
	readme.WriteString(fmt.Sprintf("# SLO Configuration: %s - %s\n\n", slo.ServiceName, slo.SLOName))
}

// writeServiceInfo writes service information section
func (g *ReadmeGenerator) writeServiceInfo(readme *strings.Builder, slo models.SLO) {
	readme.WriteString("## Service Information\n")
	readme.WriteString(fmt.Sprintf("- **Service**: %s\n", slo.ServiceName))
	readme.WriteString(fmt.Sprintf("- **SLO**: %s\n", slo.SLOName))
	readme.WriteString(fmt.Sprintf("- **Target**: %.1f%%\n", slo.Target))
	readme.WriteString(fmt.Sprintf("- **Indicator**: %s\n", slo.IndicatorType))
	readme.WriteString(fmt.Sprintf("- **Team**: %s\n", slo.Team))
	readme.WriteString(fmt.Sprintf("- **Created**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
}

// writeFilesSection writes files generated section
func (g *ReadmeGenerator) writeFilesSection(readme *strings.Builder, mainFilename string, additionalFiles map[string]string) {
	readme.WriteString("## Files Generated\n\n")
	readme.WriteString(fmt.Sprintf("### 1. Main SLO Definition\n**File**: %s\n", mainFilename))
	readme.WriteString("Contains the OpenSLO compliant SLO definition with objectives and alert policies.\n\n")

	if len(additionalFiles) > 0 {
		readme.WriteString("### 2. Alerting Configuration\n")

		if promFile, ok := additionalFiles["prometheusAlerts"]; ok {
			readme.WriteString(fmt.Sprintf("\n#### Prometheus Alert Rules\n**File**: %s\n", promFile))
			readme.WriteString("Contains Prometheus alert rules for monitoring error budget burn rate.\n")
		}

		if amFile, ok := additionalFiles["alertManagerConfig"]; ok {
			readme.WriteString(fmt.Sprintf("\n#### Alertmanager Configuration\n**File**: %s\n", amFile))
			readme.WriteString("Sample Alertmanager configuration for routing SLO alerts.\n")
		}

		if grafanaFile, ok := additionalFiles["grafanaDashboard"]; ok {
			readme.WriteString(fmt.Sprintf("\n#### Grafana Dashboard\n**File**: %s\n", grafanaFile))
			readme.WriteString("Complete Grafana dashboard for SLO monitoring.\n")
		}
	}
}

// writeAlertingSection writes alerting strategy section
func (g *ReadmeGenerator) writeAlertingSection(readme *strings.Builder, slo models.SLO) {
	readme.WriteString("\n## Alerting Strategy\n\n")
	readme.WriteString("### Burn Rate Thresholds\n")

	for _, br := range slo.Alerting.BurnRateThresholds {
		action := "Notify team, investigate within 24h"
		if br.Severity == "critical" {
			action = "Immediate action required, page on-call"
		}

		readme.WriteString(fmt.Sprintf("\n- **%s**: > %.1fx for %s\n", br.AlertName, br.Rate, br.Period))
		readme.WriteString(fmt.Sprintf("  - Severity: %s\n", br.Severity))
		readme.WriteString(fmt.Sprintf("  - Action: %s\n", action))
	}

	if len(slo.Alerting.NotificationChannels) > 0 {
		g.writeNotificationSetup(readme, slo.Alerting.NotificationChannels)
	}
}

// writeNotificationSetup writes notification setup instructions
func (g *ReadmeGenerator) writeNotificationSetup(readme *strings.Builder, channels []string) {
	readme.WriteString("\n### Notification Channels\n")
	for _, channel := range channels {
		readme.WriteString(fmt.Sprintf("- %s\n", channel))
	}

	// Add Telegram setup instructions
	for _, channel := range channels {
		if strings.HasPrefix(channel, "telegram") {
			readme.WriteString("\n### Telegram Setup Instructions\n")
			readme.WriteString("To configure Telegram notifications:\n")
			readme.WriteString("1. Set environment variables on your Alertmanager server:\n")
			readme.WriteString("   ```bash\n")
			readme.WriteString("   export TELEGRAM_BOT_TOKEN='your_bot_token_from_BotFather'\n")
			readme.WriteString("   export TELEGRAM_CHAT_ID='your_chat_id'\n")
			readme.WriteString("   ```\n")
			readme.WriteString("2. Restart Alertmanager with these environment variables\n")
			readme.WriteString("3. Ensure Alertmanager has internet access to api.telegram.org\n")
			break
		}
	}
}

// writeErrorBudgetSection writes error budget information
func (g *ReadmeGenerator) writeErrorBudgetSection(readme *strings.Builder, slo models.SLO) {
	readme.WriteString(fmt.Sprintf("\n## Error Budget\n\n"))
	readme.WriteString(fmt.Sprintf("Error Budget: %.3f%% allowable failures over %s\n", 100-slo.Target, slo.BudgetingPeriod))
}

// writeLinks writes useful links section
func (g *ReadmeGenerator) writeLinks(readme *strings.Builder) {
	readme.WriteString("\n## Useful Links\n")
	readme.WriteString("- [OpenSLO Specification](https://openslo.com)\n")
	readme.WriteString("- [SLOzy GitHub](https://github.com/philyuchkoff/slozy)\n")
	readme.WriteString("- [Alertmanager Documentation](https://prometheus.io/docs/alerting/latest/configuration/)\n")
	readme.WriteString("- [Telegram Bot API](https://core.telegram.org/bots/api)\n")
}