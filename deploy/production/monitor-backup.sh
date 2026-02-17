#!/bin/bash

# SLOzy Monitoring and Backup Script
# Place in systemd timer for automated execution

set -euo pipefail

# Configuration
DEPLOY_PATH="/opt/slozy"
BACKUP_PATH="/opt/backups/slozy"
BACKUP_RETENTION_DAYS=30
S3_BUCKET="slozy-backups-production"
PROMETHEUS_URL="http://localhost:9090"
GRAFANA_URL="http://localhost:3000"

# Function: Backup data volume
backup_data() {
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="$BACKUP_PATH/data_${timestamp}.tar.gz"

    echo "🗄️  Starting data backup..."

    # Create backup directory
    mkdir -p $BACKUP_PATH

    # Create backup
    tar -czf "$backup_file" -C "$DEPLOY_PATH/data" .

    # Upload to S3 (if AWS CLI is available)
    if command -v aws &> /dev/null; then
        aws s3 cp "$backup_file" "s3://$S3_BUCKET/data/"
        echo "✅ Backup uploaded to S3: $(basename $backup_file)"
    fi

    # Verify backup integrity
    if tar -tzf "$backup_file" &> /dev/null; then
        echo "✅ Backup verified: $backup_file"
    else
        echo "❌ Backup verification failed!"
        exit 1
    fi

    # Cleanup old backups
    find $BACKUP_PATH -name "data_*.tar.gz" -mtime +$BACKUP_RETENTION_DAYS -delete
}

# Function: Check service health
check_health() {
    local health_url="http://localhost:8080/health"
    local response=$(curl -s "$health_url" || echo "")

    if [[ $response == *"healthy"* ]]; then
        echo "✅ Health check passed"
        return 0
    else
        echo "❌ Health check failed"
        return 1
    fi
}

# Function: Check security metrics
check_security() {
    echo "🔒 Analyzing security logs..."

    # Check for injection attempts
    local injections=$(journalctl -u slozy --since "1 hour ago" | grep -ic '(\.\./|\.\.\\|%3e|%2f)' || echo "0")
    if [ "$injections" -gt 0 ]; then
        echo "⚠️  Path traversal attempts: $injections"
    fi

    # Check for rate limit violations
    local rate_limits=$(journalctl -u slozy --since "1 hour ago" | grep -ic 'rate limit' || echo "0")
    if [ "$rate_limits" -gt 0 ]; then
        echo "📈 Rate limit violations: $rate_limits"
    fi

    # Check for blocked requests
    local blocked=$(journalctl -u slozy --since "1 hour ago" | grep -ic '(blocked|forbidden|unauthorized)' || echo "0")
    if [ "$blocked" -gt 10 ]; then
        echo "🚨 High number of blocked requests: $blocked"
    fi
}

# Function: Extract Prometheus metrics
check_metrics() {
    if ! curl -s "$PROMETHEUS_URL/api/v1/query?query=up" | grep -q '"value":\[1\]'; then
        echo "⚠️  Prometheus not reporting healthy metrics"
        return 1
    fi

    # Get request rate
    local request_rate=$(curl -s "$PROMETHEUS_URL/api/v1/query?query=rate(http_requests_total\[5m\])" | jq -r '.data.result[0].value[1] || 0')
    echo "📊 Request rate: $request_rate req/s"

    # Check error rate
    local error_rate=$(curl -s "$PROMETHEUS_URL/api/v1/query?query=rate(http_requests_total{status=~\"5..\"}\[5m\])" | jq -r '.data.result[0].value[1] || 0')
    if (( $(echo "$error_rate > 0.1" | bc -l) )); then
        echo "🚨 High error rate: $error_rate"
    fi
}

# Function: Deploy update
deploy_update() {
    echo "🔄 Deploying update..."

    # Stop current service
    systemctl stop slozy || docker-compose down

    # Backup before update
    backup_data

    # Deploy new version
    ./deploy/production/deploy.sh --domain $DOMAIN_NAME

    # Verify deployment
    if check_health; then
        echo "✅ Update deployed successfully"
        # Notify monitoring system
        curl -X POST "$PROMETHEUS_URL/api/v1/alerts" -d @- <<EOF
[{
  "labels": {
    "alertname": "SLOzyUpdate",
    "service": "slozy",
    "severity": "info"
  },
  "annotations": {
    "summary": "SLOzy updated successfully",
    "description": "Version deployed at $(date)"
  }
}]
EOF
    else
        echo "❌ Update failed, rolling back..."
        systemctl restart slozy || docker-compose up
    fi
}

# Function: Generate security report
generate_security_report() {
    local report_file="$BACKUP_PATH/security_report_$(date +%Y%m%d).json"
    local timestamp=$(date -Iseconds)

    echo "📋 Generating security report..."

    # Collect logs from last 24 hours
    local logs=$(journalctl -u slozy --since "24 hours ago" --no-pager -o json)

    # Parse logs and generate report
    echo {} | jq "
        .timestamp = \"$timestamp\" |
        .period = \"24h\" |
        .stats = {
            total_requests: $(echo \"$logs\" | jq -s 'length'),
            path_traversal_attempts: $(echo \"$logs\" | jq -r 'select(.MESSAGE | test(\"\\.\\./|\\.\\.\\\\|%3e|%2f\"; \"i\")) | length'),
            rate_limit_hits: $(echo \"$logs\" | jq -r 'select(.MESSAGE | test(\"rate limit\"; \"i\")) | length'),
            blocked_requests: $(echo \"$logs\" | jq -r 'select(.MESSAGE | test(\"blocked|forbidden|unauthorized\"; \"i\")) | length'),
            errors_4xx: $(echo \"$logs\" | jq -r 'select(.SYSTEMD_UNIT == \"slozy\" and (_STREAM == \"stdout\" or _STREAM == \"stderr\") | select(.MESSAGE | test(\"HTTP 4[0-9][0-9]\")) | length'),
            errors_5xx: $(echo \"$logs\" | jq -r 'select(.SYSTEMD_UNIT == \"slozy\" and (_STREAM == \"stdout\" or _STREAM == \"stderr\") | select(.MESSAGE | test(\"HTTP 5[0-9][0-9]\")) | length')
        }
    " > "$report_file"

    echo "✅ Security report saved: $report_file"
}

# Function: Cleanup old data
cleanup() {
    echo "🧹 Performing cleanup..."

    # Rotate logs
    find "$DEPLOY_PATH/logs" -name "*.log" -mtime +7 -exec mv {} {}.old \;
    find "$DEPLOY_PATH/logs" -name "*.log.old" -mtime +30 -delete

    # Clean temp files
    find /tmp -name "slozy_*" -mtime +1 -delete

    # Cleanup Docker images if using Docker
    if command -v docker &> /dev/null; then
        docker system prune -f --volumes
    fi

    echo "✅ Cleanup complete"
}

# Function: Send alerts
send_alert() {
    local severity=$1
    local message=$2
    local details=$3

    # Send to Grafana
    if [ -n "$GRAFANA_WEBHOOK" ]; then
        curl -X POST "$GRAFANA_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d "{
                \"severity\": \"$severity\",
                \"message\": \"$message\",
                \"details\": $details,
                \"timestamp\": \"$(date -Iseconds)\",
                \"service\": \"slozy\"
            }"
    fi

    # Log to journal
    logger -t slozy-monitor "[ALERT] $message"
}

# Main execution
case ${1:-health} in
    health)
        check_health
        ;;
    backup)
        backup_data
        ;;
    security)
        check_security
        generate_security_report
        ;;
    metrics)
        check_metrics
        ;;
    update)
        deploy_update
        ;;
    cleanup)
        cleanup
        ;;
    all)
        echo "🔍 Running full monitoring suite..."
        echo ""
        check_health
        check_security
        check_metrics
        echo ""
        echo "✅ Monitoring complete"
        ;;
    alert)
        if [ -z "${2:-}" ]; then
            echo "Usage: $0 alert severity 'message'"
            exit 1
        fi
        send_alert "$2" "${3:-}" "${4:-}"
        ;;
    *)
        echo "Usage: $0 {health|backup|security|metrics|update|cleanup|all|alert severity message}"
        echo ""
        echo "Commands:"
        echo "  health    - Check service health"
        echo "  backup    - Backup data volume"
        echo "  security  - Analyze security logs"
        echo "  metrics   - Extract Prometheus metrics"
        echo "  update    - Deploy update with rollback"
        echo "  cleanup   - Rotate logs and cleanup"
        echo "  all       - Run full monitoring suite"
        echo "  alert    - Send alert message"
        exit 1
        ;;
esac