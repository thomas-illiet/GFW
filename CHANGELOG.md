# Changelog

All notable changes to Great Firewall are documented here. Application releases
follow semantic versioning; the HTTP API has its own independently versioned
contract in `openapi.yaml`.

## 0.1.0 - 2026-10-29

### Added

- PostgreSQL-backed ecosystem, machine, trusted-service and ingestion-key management.
- Deterministic firewall lists with ETags and request history dashboards.
- OIDC administration portal with Viewer, Editor and Admin roles.
- Forward-only, embedded Goose migrations with PostgreSQL session locking.
- Rootless container, Docker Compose environment and production-oriented Helm chart.

### Operations

- This is the first supported database schema. Databases created by pre-release
  builds must be deleted and recreated before installing 0.1.0.
