# Great Firewall

Great Firewall is an IPv4 policy control plane. It maintains machine addresses per product ecosystem, resolves globally trusted service hostnames, and publishes deterministic JSON or text lists for firewall automation.

## Features

- PostgreSQL-backed machine, ecosystem and trusted-service catalogues.
- `GET /api/v1/firewall/ecosystems/{ecosystem}/ips?format=json|txt` with stable ETags.
- Detailed one-year request history and a 24-hour, 7-day and 30-day activity dashboard.
- OIDC administration with Viewer, Editor and Admin roles.
- Revocable, hashed API keys for managed inventory ingestion.
- Background DNS refresh that retains the last successful IPv4 set on failure.
- Nuxt/Vuetify administration portal embedded in the production image.

## Local development

The quickest complete environment uses development authentication:

```sh
docker compose up --build
```

Open <http://localhost:8080>. Development authentication grants the local user the Admin role and must never be enabled in production.

To populate the local database with an idempotent 30-day demonstration dataset:

```sh
make demo-seed
```

The dataset adds product ecosystems, machines, trusted services, ingestion keys and detailed firewall request history. It uses documentation-only IPv4 ranges and must not be loaded in production.

For native development, start PostgreSQL and set:

```sh
export GFW_DATABASE_URL='postgres://gfw:gfw@localhost:5432/gfw?sslmode=disable'
export GFW_DEV_AUTH=true
make frontend-build
make migrate
make run
```

## Production configuration

| Variable | Purpose | Default |
| --- | --- | --- |
| `GFW_DATABASE_URL` | PostgreSQL connection URL | local development URL |
| `GFW_PUBLIC_URL` | Browser-visible base URL | `http://localhost:8080` |
| `GFW_OIDC_ISSUER` | OpenID Connect issuer | required |
| `GFW_OIDC_CLIENT_ID` | OIDC client identifier | required |
| `GFW_OIDC_CLIENT_SECRET` | OIDC client secret | required |
| `GFW_OIDC_REDIRECT_URL` | OIDC callback URL | `{public URL}/auth/callback` |
| `GFW_SESSION_SECRET` | At least 32 bytes used to sign sessions | required |
| `GFW_DNS_INTERVAL` | Trusted DNS refresh interval | `15m` |
| `GFW_DNS_TIMEOUT` | Per-host resolution timeout | `5s` |
| `GFW_RETENTION_DAYS` | Detailed request retention | `365` |

OIDC groups `read`, `write`, and `admin` map to Viewer, Editor, and Admin.
The prefixed forms `gfw-read`, `gfw-write`, and `gfw-admin` are also accepted,
as are the legacy `viewer` and `editor` names. Roles are cumulative and Admin
takes precedence when a user belongs to several groups. Authenticated users
without a recognized group receive read-only Viewer access. See the
[authentication and authorization guide](docs/authentication.md) for the exact
permissions.

## Ingestion

Create a key in the Administration UI, then submit partial upsert batches:

```sh
curl -H 'Authorization: Bearer PREFIX.SECRET' \
  -H 'Content-Type: application/json' \
  --data '{"machines":[{"hostname":"api-01","ecosystemID":"ECOSYSTEM_UUID","ipv4":"10.0.0.10"}]}' \
  http://localhost:8080/api/v1/ingestion/machines:upsert
```

See [`openapi.yaml`](openapi.yaml) for the HTTP contract.

An adaptable Python example is available in
[`scripts/ingest_source.py`](scripts/ingest_source.py). It reads a Source X
CSV export from a file or HTTP endpoint, validates IPv4 addresses, and sends
bounded batches to the ingestion API without third-party dependencies. The CSV
must contain the columns `machine`, `ecosystem`, and `ip`, as shown in
[`scripts/source_x.example.csv`](scripts/source_x.example.csv). Test the
transformation with (comma, semicolon, and tab delimiters are accepted):

```sh
python3 scripts/ingest_source.py --dry-run
```

The referenced ecosystems must already exist in Great Firewall. Map the readable
CSV values to their IDs in
[`scripts/ecosystem_map.example.json`](scripts/ecosystem_map.example.json), then
ingest a real HTTP source with:

```sh
export SOURCE_X_URL='https://inventory.example.test/api/assets'
export SOURCE_X_TOKEN='source-x-token'
export GFW_URL='https://gfw.example.test'
export GFW_INGESTION_KEY='PREFIX.SECRET'
export GFW_ECOSYSTEM_MAP='scripts/ecosystem_map.example.json'
python3 scripts/ingest_source.py
```

Adapt `load_source_x` if Source X uses a different authentication method. The
ingestion operation is an upsert: sending the same ecosystem/IP pair refreshes
its machine name and synchronization timestamp.

## Backend commands

`gfw serve` starts the HTTP service, `gfw migrate up` applies all pending
forward-only migrations, `gfw migrate status` and `gfw migrate version` inspect
the schema, and `gfw version` prints build metadata. `gfw migrate` is an alias
for `gfw migrate up`. Running `gfw` without a subcommand is equivalent to
`gfw serve`; the server never changes the schema.

Configuration can come from YAML (`--config`), `GFW_*` environment variables, or flags. Start with the [technical documentation index](docs/README.md), then see the [configuration guide](docs/configuration.md) and [architecture documentation](docs/architecture.md).

Run `make coverage-check` to enforce the backend's 80% coverage floor, or `make verify` for the complete backend and frontend quality suite.

## Kubernetes / Helm

The production-oriented chart in [`charts/gfw`](charts/gfw) deploys the server,
schema migration, Service and optional Ingress without installing PostgreSQL.
The database URL and OIDC/session secrets are injected from a Kubernetes Secret.
The chart enforces UID/GID 1000,
read-only root filesystems, dropped capabilities, no privilege escalation,
`RuntimeDefault` seccomp, disabled service-account token mounting and a bounded
`/tmp` volume. See the [chart README](charts/gfw/README.md) for installation.

## Manual release

Application releases use semantic versioning. The OpenAPI contract remains at
`1.0.0` and is versioned independently. Run `make release-check VERSION=0.1.0`
to execute the complete suite, build and smoke-test the versioned image, and
package the Helm chart under `dist/`. No artifact is published automatically.

Version 0.1.0 is the first supported schema. If a database was created by a
pre-release build, recreate it first (`docker compose down --volumes` for the
local environment); automatic baselining is intentionally unsupported.
