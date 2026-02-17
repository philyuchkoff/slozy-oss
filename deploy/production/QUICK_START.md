# Quick Start Guide for Production Deployment

## 🚀 Fast Track Deployment (for testing/demo)

### 1. Docker Compose (Easiest)

```bash
# 1. Clone and navigate
git clone https://github.com/philyuchkoff/slozy.git
cd slozy/deploy/production

# 2. Configure environment
cp .env.example .env
# Edit .env with your settings:
# DOMAIN=yourdomain.com
# TELEGRAM_BOT_TOKEN=your_token
# TELEGRAM_CHAT_ID=your_id

# 3. Deploy
docker-compose -f docker-compose.prod.yml up -d

# 4. Verify
curl -f http://localhost/health
```

### 2. Binary Deployment

```bash
# 1. Build
go run deploy/production/build.sh

# 2. Deploy with script
sudo ./deploy/production/deploy.sh --domain yourdomain.com

# 3. Setup reverse proxy
# Configure Nginx or use provided config
```

## 🔒 Security Verification Checklist

After deployment, run these commands:

```bash
# 1. Check security headers
curl -I https://yourdomain.com/health | grep -E "(X-Frame|X-Content|CSP)"

# 2. Test rate limiting
for i in {1..6}; do
    curl -s https://yourdomain.com/api/slos
    echo "Request $i: $(date)"
    sleep 0.2
done

# 3. Test path traversal
curl -s -X POST https://yourdomain.com/api/slo/create \
    -H "Content-Type: application/json" \
    -d '{"serviceName":"../../../etc/passwd","sloName":"test","target":99.9}'

# 4. Verify file restrictions
curl -I https://yourdomain.com/data/../config.go
# Should return 403/404

# 5. Test CORS
curl -H "Origin: https://evil.com" \
     -I https://yourdomain.com/api/slos
# Should not return Access-Control-Allow-Origin: https://evil.com
```

## 📊 Monitoring Setup

### Enable Monitoring (Docker)

1. Access Grafana: http://localhost/grafana
   - User: admin
   - Pass: change_me (immediately!)

2. Access Prometheus: http://localhost/grafana/prometheus
   - View metrics at http://localhost/grafana/prometheus/targets

3. Key metrics to monitor:
   - SLOzy request rate
   - Error rate (should be <1%)
   - Response time (95th percentile)
   - Rate limit violations

### Production Monitoring Alerts

Set up alerts for:
- 5xx errors > 1/s for 5 minutes
- Rate limit violations > 10/minute
- Health check failures
- High memory/CPU usage

🎉 You're all set! Your SLOzy instance is secure and production-ready.

## Need Help?

1. Check logs: `journalctl -u slozy -f` or `docker-compose logs -f`
2. Restart service: `systemctl restart slozy` or `docker-compose restart`
3. Test locally first with `make test`
4. Review logs for security events

## Configuration Examples

### High-Traffic Site (1000+ req/s)

```bash
# .env
API_RATE_LIMIT=100
CREATE_SLO_RATE_LIMIT=20
# Add more app instances behind load balancer
```

### Private Network

```bash
# .env
ALLOWED_ORIGINS=https://internal.corp.local,https://corp-dashboard.local
# Configure VPN access only
```

### Compliance Mode

```bash
# nginx.conf additions
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'nonce-$nonce'" always;
add_header X-Content-Type-Options: nosniff always;
# Enable audit logging
access_log /var/log/nginx/slozy-audit.log audit;