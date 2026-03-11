# SLOzy v1.2.1 Release Notes

## 📋 Overview
SLOzy v1.2.1 is a **security-focused release** that addresses critical vulnerabilities discovered in the previous version. While no new features are introduced, this release contains essential security patches that significantly strengthen the application's security posture.

## 🚨 Critical Security Fixes

### 1. CORS Misconfiguration Fix
- **Issue**: Hardcoded CORS headers in multiple handlers allowed any origin
- **Fix**: Centralized CORS configuration through middleware
- **Impact**: Prevents unauthorized cross-origin requests

### 2. Content Security Policy (CSP) Strengthening
- **Issue**: Policy allowed `unsafe-inline` and `unsafe-eval`, enabling XSS attacks
- **Fix**: Removed unsafe directives from CSP
- **Impact**: Prevents XSS attacks through injection prevention

### 3. Prometheus Query Injection Protection
- **Issue**: User input directly inserted into PromQL without validation
- **Fix**:
  - Character whitelist validation
  - Dangerous pattern detection
  - Query sanitization
- **Impact**: Prevents command injection and remote code execution

## 🔒 Security Improvements

### Input Validation
- Added comprehensive character validation for Prometheus queries
- Implemented dangerous keyword detection (exec, system, shell, etc.)
- Added command substitution prevention

### CORS Configuration
- All CORS handling now centralized in middleware
- Configurable allowed origins
- Removed unsafe wildcard settings

### Content Security Policy
```diff
- default-src 'self' 'unsafe-inline' 'unsafe-eval';
+ default-src 'self';
```

## 📊 Affected Components

| Component | Severity | Status |
|-----------|-----------|--------|
| API Handlers | High | ✅ Fixed |
| Middleware | High | ✅ Fixed |
| Prometheus Generator | Critical | ✅ Fixed |
| Validation Module | High | ✅ Enhanced |

## 🎯 Testing Recommendations

After upgrading to v1.2.1, perform the following tests:

1. **CORS Testing**
   ```bash
   curl -H "Origin: https://evil.com" -X OPTIONS http://localhost:8080/api/slo/create
   ```
   Expected: No Access-Control-Allow-Origin header

2. **XSS Testing**
   - Attempt to inject inline scripts in the web interface
   - Verify CSP blocks execution

3. **SQL/PromQL Injection Testing**
   ```bash
   curl -X POST http://localhost:8080/api/slo/create \
     -H "Content-Type: application/json" \
     -d '{"metricQuery": "rate(...) AND exec(\"whoami\")"}'
   ```
   Expected: Validation error

## 🔄 Migration Guide

No breaking changes introduced. Simply update your deployment:

```bash
# If using Docker
docker pull slozy:latest

# If using Git
git pull origin main
make build
./bin/slozy

# If using deploy script
sudo ./deploy/install.sh
```

## 🙏 Acknowledgments

Special thanks to the security review team for identifying these vulnerabilities:
- CORS misconfiguration audit team
- Web security penetration testers
- Prometheus security experts

## 🔗 Resources

- [Security Fixes Details](SECURITY_FIXES.md)
- [Complete Changelog](CHANGELOG.md)
- [Security Policy](./SECURITY.md)

## ⚠️ Important Note

This release addresses **critical security vulnerabilities**. **Immediate upgrade is strongly recommended** for all production deployments.

The vulnerabilities addressed could potentially allow:
- Cross-origin data theft
- XSS attacks on users
- Command execution through Prometheus queries
- Remote code execution in monitoring systems

---

**Release Date**: March 11, 2026
**Version**: 1.2.1
**Type**: Security Patch