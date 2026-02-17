Security Fixes and Production Deployment Infrastructure

## Summary

This commit implements comprehensive security fixes to address critical vulnerabilities identified in a security audit, and provides complete production deployment infrastructure.

### Critical Security Fixes 🛡️

1. **Path Traversal Protection**
   - Fixed safeFilename to sanitize directory traversal attempts
   - Added Unicode normalization (golang.org/x/text/unicode/norm)
   - Implemented Windows reserved filename checking
   - Added null byte handling and length limits

2. **Secure Data Endpoint**
   - Replaced unrestricted file server with secure handler
   - Only allows .yaml file access
   - Validates paths stay within data directory
   - Prevents directory traversal attacks

3. **Security Headers Middleware** (internal/middleware/security.go)
   - X-Content-Type-Options: nosniff
   - X-Frame-Options: DENY
   - X-XSS-Protection: 1; mode=block
   - Content-Security-Policy: default-src 'self'
   - Strict-Transport-Security (HTTPS only)

4. **CORS Protection**
   - Replaced wildcard (*) with configurable origins
   - Added OPTIONS method handling
   - Applied only to API endpoints

5. **Rate Limiting** (internal/middleware/ratelimit.go)
   - Implemented token bucket algorithm (golang.org/x/time/rate)
   - 10 rps default, 5 rps for SLO creation
   - IP-based tracking with automatic cleanup
   - Per-endpoint configuration support

6. **Input Validation Enhancements**
   - Added maximum length checks for all fields
   - Service/SLO names: 100 chars max
   - Metric queries: 10,000 chars max
   - General fields: 1,000 chars max

7. **Request Size Limiting** (internal/middleware/validation.go)
   - 1MB maximum for request bodies
   - Early validation prevents DoS
   - Returns 413 for oversized requests

8. **ZIP Filename Sanitization**
   - Clean all filenames before ZIP creation
   - Remove path separators
   - Validate against dangerous characters
   - Ensure relative paths only

### Production Infrastructure 🚀

1. **Deployment Scripts**
   - build.sh: Multi-platform binary creation
   - deploy.sh: Automated production deployment
   - monitor-backup.sh: Monitoring and backup utilities

2. **Docker Support**
   - Production Dockerfile with security hardening
   - Docker Compose with monitoring stack
   - Multi-stage builds for optimized images

3. **Infrastructure as Code**
   - AWS Terraform configuration
   - Nginx reverse proxy with security headers
   - Cloud-init for automated EC2 deployment

4. **Monitoring & Observability**
   - Prometheus metrics configuration
   - Grafana dashboards
   - Log aggregation (systemd/journal)
   - Security event tracking

### Test Coverage 🧪

- Security headers middleware tests
- CORS middleware tests
- Input validation tests
- Filename sanitization tests
- SLO handler security tests
- Path traversal prevention tests

### Breaking Changes ⚠️

- CORS configuration now requires explicit origins
- Default rate limits may require adjustment for high-traffic sites
- Some filenames with special characters will be modified

### Migration Steps 📋

1. Run tests: `go test ./...`
2. Update configuration with new environment variables
3. Set ALLOWED_ORIGINS for your domain(s)
4. Adjust rate limits if needed
5. Configure reverse proxy for HTTPS

### Added Dependencies 📦

- golang.org/x/time/rate: For rate limiting
- golang.org/x/text/unicode/norm: For Unicode normalization

### Files Added 📁

- internal/middleware/security.go
- internal/middleware/ratelimit.go
- internal/middleware/validation.go
- internal/middleware/security_test.go
- internal/models/validator_test.go
- internal/handlers/slo_test.go
- deploy/production/* (complete deployment package)
- SECURITY_FIXES.md (documentation)
- test_security_fixes.sh (verification script)
- verify_security.sh (quick validation)

### Performance Impact 📊

- Minimal overhead from security middleware
- Rate limiting prevents abuse
- Input validation reduces processing time for invalid requests
- Caching in Nginx improves static asset delivery

### Security Improvements 🔒

This update mitigates:
- Directory traversal attacks
- XSS vulnerabilities
- Clickjacking attempts
- DoS attacks via large payloads
- CORS-based attacks
- Information disclosure via file access

## Testing

All unit tests pass:
```
✅ Security Headers Tests: PASS
✅ CORS Tests: PASS
✅ Validator Tests: PASS
✅ Safe Filename Tests: PASS
✅ SLO Handler Tests: PASS
```

## Verification

Run security verification:
```bash
./test_security_fixes.sh  # Full integration tests
./verify_security.sh  # Quick validation
```

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>