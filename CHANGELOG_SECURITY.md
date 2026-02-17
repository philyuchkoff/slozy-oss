# Security Changelog

All notable changes to this project are documented in this file.

## [1.2.0] - 2026-02-17

### 🛡️ Security

This release implements comprehensive security fixes for all vulnerabilities identified in the security audit. **AFFECTED USERS SHOULD UPDATE IMMEDIATELY.**

#### Added

- **Path Traversal Protection**
  - Sanitize all user input with Unicode normalization
  - Protect against directory traversal attacks (../)
  - Windows reserved filename checking
  - Null byte removal
  - Length limits on all string fields

- **Security Headers Middleware**
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - X-XSS-Protection: 1; mode=block
  - Referrer-Policy: strict-origin-when-cross-origin
  - Content-Security-Policy: default-src 'self'
  - Strict-Transport-Security (HTTPS only)

- **CORS Protection**
  - Configurable allowed origins (no wildcards)
  - Proper OPTIONS method handling
  - Credentials control
  - Endpoint-specific configuration

- **Rate Limiting**
  - Token bucket algorithm implementation
  - Default: 10 req/s, burst 20
  - SLO creation: 5 req/s, burst 10
  - IP-based tracking with cleanup
  - Memory efficient implementation

- **Request Size Validation**
  - 1MB maximum request body
  - Early rejection of oversized payloads
  - HTTP 413 status codes
  - Memory protection against DoS

- **Secure File Access**
  - /data/ endpoint restricted to .yaml files
  - Path validation ensuring directory confinement
  - Template handler path verification
  - Symbolic link resolution

- **Input Validation Enhancements**
  - Service name: max 100 chars
  - SLO name: max 100 chars
  - Metric query: max 10,000 chars
  - General fields: max 1,000 chars
  - Early validation for all endpoints

- **ZIP Archive Security**
  - Filename sanitization before creation
  - Path separator removal
  - Relative paths enforced
  - Safe character filtering

### Changed

- CORS now requires explicit domain configuration
- All security middleware applied via composition
- Error messages sanitized to prevent information disclosure

### Security

#### Critical
- CVE-2024-XXXX: Path Traversal in Filename Generation - FIXED
- CVE-2024-YYYY: Unrestricted File Access - FIXED
- CVE-2024-ZZZZ: CORS Misconfiguration - FIXED
- CVE-2024-WWWW: Missing Security Headers - FIXED

#### High
- DoS via Large Payload - FIXED
- Directory Traversal via ZIP - FIXED
- Missing Input Validation - FIXED

#### Medium
- Rate Limiting Absence - FIXED
- Information Disclosure - FIXED

### Added
- `golang.org/x/time/rate` - Rate limiting
- `golang.org/x/text/unicode/norm` - Unicode normalization

### Production

- Complete deployment infrastructure added
- Docker hardening configurations
- TLS/SSL automation scripts
- Monitoring and alerting setup
- Backup and disaster recovery
- Infrastructure as code (Terraform)

## [1.1.0] - Previous Release (Insecure)

❌ **This version should NOT be used in production**

### Vulnerabilities Present

- Path traversal via service/SLO names
- Unrestricted file access at /data/
- No security headers
- Wildcard CORS (*)
- No rate limiting
- No input size validation
- Missing input validation
- ZIP archives with malicious paths

## Upgrade Instructions

### From 1.1.0 to 1.2.0

1. **Immediate Actions Required:**
   ```bash
   git checkout v1.2.0
   go mod tidy
   go build .
   ```

2. **Configuration Changes:**
   - Add `ALLOWED_ORIGINS` to environment
   - Set `API_RATE_LIMIT` (default: 10)
   - Set `CREATE_SLO_RATE_LIMIT` (default: 5)
   - Enable TLS certificates

3. **Deployment:**
   ```bash
   # Using new secure deployment
   ./deploy/production/deploy.sh --domain yourdomain.com

   # Or Docker
   docker-compose -f deploy/production/docker-compose.prod.yml up -d
   ```

### Verification

After upgrade, verify security:
```bash
# Check headers
curl -I https://yourdomain.com/api/slo/create

# Test rate limiting
for i in {1..20}; do curl -sO /dev/null https://yourdomain.com/api/slos; done

# Verify path protection
curl -X POST https://yourdomain.com/api/slo/create \
    -d '{"serviceName":"../../../etc/passwd","sloName":"test"}'
```

## Security Policy

This project follows responsible disclosure:
- Security issues are fixed in next release
- CVE numbers are registered for critical issues
- Security advisory email: security@slozy.com
- Public disclosure after patch release

### Supported Versions

| Version | Security Support | End of Life |
|---------|-------------------|--------------|
| 1.2.0   | ✅ Yes           | TBD          |
| < 1.2.0 | ❌ No (upgrade required) | Immediately |

### Reporting Security Issues

If you discover a security vulnerability:
1. Do NOT open a public issue
2. Email details to: security@slozy.com
3. Include reproduction steps
4. We'll respond within 48 hours
5. Fix will be released within 7 days

## Acknowledgments

Security audit performed by:
- Internal security team
- External penetration testing
- Community disclosure
- Static analysis tools

Thanks to all security researchers who contributed to making SLOzy more secure.