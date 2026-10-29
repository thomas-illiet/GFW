# Testing strategy

Backend tests are deterministic and require neither Docker nor an external PostgreSQL instance. HTTP and DNS tests use narrow fake repositories; GORM repository tests use `go-sqlmock` to validate generated SQL, transactions, failures, and PostgreSQL-specific paths.

`make coverage-check` measures all packages under `internal/...`, writes text and HTML reports under `coverage/`, and fails below 80%. The executable bootstrap remains outside the threshold, while the Cobra command package is included and directly tested.

`make verify` additionally runs formatting checks, `go vet`, documentation and
size guards, the race detector, frontend checks, OpenAPI route validation, Helm
lint/rendering, Compose validation and release-version consistency checks.

`make release-check VERSION=0.1.0` also builds and smoke-tests the final image
and packages the chart. Important scenarios include configuration precedence,
forward-only migration state, legacy-schema rejection, missing resources,
authorization roles, CSRF, ETags, strict JSON bodies, DNS failure retention,
worker cancellation and graceful HTTP behavior.
