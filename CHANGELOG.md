# Changelog

All notable changes to this project will be documented in this file.

## [1.4.1] - 2026-07-27

### Changed
- Bump Go toolchain floor to 1.26.5 to clear five called standard-library CVEs.
- Resolve linter and modernize findings (errcheck, G404 annotation, min/max builtins, strings.Cut); no behavior change.
- Sync README with current behavior: IMEI validation, QueryResult JSON tags, and both retry layers.
