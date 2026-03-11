# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.1] - 2026-03-11

### Security
- Fixed critical CORS misconfiguration by removing hardcoded headers from handlers
- Strengthened Content Security Policy by removing unsafe-inline and unsafe-eval
- Implemented comprehensive Prometheus query injection protection
- Added character whitelist validation for PromQL queries
- Added dangerous pattern detection (exec, system, shell, etc.)
- Added query sanitization before inclusion in expressions

### Changed
- All CORS configurations now centralized in middleware
- Enhanced input validation for PromQL queries
- Improved defense-in-depth security measures

### Fixed
- Prevented potential XSS attacks through CSP bypass
- Fixed command injection vulnerability in Prometheus queries
- Resolved remote code execution risks in monitoring systems

## [1.2.0] - 2026-02-20

### Added
- Comprehensive security headers middleware
- Rate limiting using token bucket algorithm
- Input validation with length limits
- Request size limiting (1MB max)
- ZIP filename sanitization
- Path traversal protection with Unicode normalization
- Secure data endpoint (.yaml files only)

### Security
- CORS protection with configurable origins
- HSTS support
- XSS protection headers
- Frame options protection
- Content-Type options protection

## [1.1.0] - 2024-12-18

### Added
- OpenSLO compliance support
- Template management system
- ZIP file download functionality
- Web UI improvements

## [1.0.0] - 2024-11-15

### Added
- Initial release
- SLO creation and management
- Prometheus alert generation
- Grafana dashboard generation
- AlertManager configuration
- Basic web interface
- API endpoints