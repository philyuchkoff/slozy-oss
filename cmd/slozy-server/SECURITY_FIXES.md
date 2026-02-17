# SLOzy Security Fixes Documentation

## Overview

This document outlines the security vulnerabilities discovered and fixed in the SLOzy application to ensure a robust and secure deployment.

## Security Issues Fixed

### 1. Path Traversal in Filename Generation

**Issue**: The `safeFilename` function allowed directory traversal by not properly sanitizing input strings.

**Fix**:
- Added Unicode normalization using `golang.org/x/text/unicode/norm`
- Implemented `filepath.Base()` and `filepath.Clean()` for path sanitization
- Added Windows reserved filename checks
- Implemented length limits (max 100 chars per component)
- Added null byte handling

**File**: `internal/handlers/slo.go`

### 2. Unrestricted /data/ Endpoint

**Issue**: Direct file server exposed all files in the data directory to anyone.

**Fix**:
- Replaced simple file server with secure handler
- Only allows `.yaml` file access
- Validates paths to ensure files stay within data directory
- Prevents directory traversal attempts
- Returns 404/403 for unauthorized access

**File**: `main.go`

### 3. ZIP Archive Directory Traversal

**Issue**: ZIP creation could include malicious file paths.

**Fix**:
- Added filename sanitization before ZIP creation
- Removed path separators to ensure relative paths only
- Validated filenames against dangerous characters
- Ensured maximum filename length (255 chars)

**File**: `internal/handlers/slo.go`

### 4. Missing Security Headers

**Fix**: Implemented security headers middleware with:
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- X-XSS-Protection: 1; mode=block
- Referrer-Policy: strict-origin-when-cross-origin
- Content-Security-Policy: default-src 'self' with appropriate exemptions
- Strict-Transport-Security (HTTPS only)

**File**: `internal/middleware/security.go`

### 5. Weak CORS Configuration

**Issue**: Wildcard CORS allowed all origins.

**Fix**:
- Configurable allowed origins list
- Development-friendly localhost wildcard
- Proper OPTIONS method handling
- Applied only to API endpoints

**File**: `internal/middleware/security.go`

### 6. No Rate Limiting

**Fix**: Implemented token bucket rate limiting:
- Default: 10 requests/second, burst of 20
- Create SLO endpoint: 5 requests/second, burst of 10
- IP-based tracking with automatic cleanup
- Memory-efficient implementation

**File**: `internal/middleware/ratelimit.go`

### 7. Unlimited Request Sizes

**Fix**: Request size validation:
- Default limit: 1MB for body content
- Applies to POST/PUT/PATCH requests
- Returns 413 for oversized requests
- Early detection to prevent resource consumption

**File**: `internal/middleware/validation.go`

### 8. Input Length Validation

**Fix**: Enhanced validation with length checks:
- Service Name: max 100 chars
- SLO Name: max 100 chars
- Team: max 100 chars
- Threshold: max 50 chars
- Metric Query: max 10,000 chars
- All fields: general max of 1,000 chars

**File**: `internal/models/validator.go`

### 9. Template Handler Path Traversal

**Fix**: Added path validation to `loadTemplateFile`:
 Ensures accessed files are within templates directory
- Only allows `.yaml` files
- Resolves symbolic links to prevent bypass
- Returns errors for unauthorized access attempts

**File**: `internal/handlers/template.go`

## Security Testing

A comprehensive security test script has been provided at `test_security_fixes.sh`. This script tests:

1. Path traversal protection in SLO creation
2. Large payload DoS protection
3. Data endpoint security
4. Security headers presence
5. Rate limiting effectiveness
6. ZIP filename sanitization
7. CORS protection

## Usage

```bash
# Run the server
cd cmd/slozy-server
go run .

# In another terminal, run security tests
./test_security_fixes.sh
```

## Dependencies Added

- `golang.org/x/time/rate` - For rate limiting
- `golang.org/x/text/unicode/norm` - For Unicode normalization

## Configuration

### Custom Rate Limits

You can customize rate limits by modifying the configuration in `main.go`:

```go
apiRateLimit := &middleware.RateLimitConfig{
    BurstSizeBPS:    10,  // requests per second
    BurstSizeMAX:    20,  // burst size
    CleanupInterval: time.Minute,
}
```

### Custom CORS Origins

Update the CORS configuration in `main.go`:

```go
corsConfig := &middleware.CORSConfig{
    AllowedOrigins: []string{"https://yourdomain.com"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
}
```

## Monitoring and Alerting

1. **Rate Limited Requests**: Monitor logs for 429 status codes
2. **Blocked Path Traversals**: Look for 400/403 errors on `/data/` endpoint
3. **Security Headers**: Verify headers are present in all responses
4. **Large Payloads**: Monitor for 413 status codes

## Best Practices

1. **Deploy with HTTPS**: Ensure HSTS header is active
2. **Monitor Rate Limits**: Adjust limits based on usage patterns
3. Regular Updates: Keep dependencies updated
4. **Security Reviews**: Conduct regular security audits
5. **Error Handling**: Ensure sensitive information doesn't leak in error messages

## Post-Deployment Checklist

- [ ] Verify all security headers are present
- [ ] Test rate limiting is working
- [ ] Confirm path traversal is blocked
- [ ] Check CORS is properly configured
- [ ] Validate request size limits
- [ ] Run comprehensive security test suite
- [ ] Monitor error logs for blocked attacks
- [ ] Update documentation if needed

## Contact

For security concerns or vulnerabilities, please submit a security report through your organization's vulnerability disclosure process.