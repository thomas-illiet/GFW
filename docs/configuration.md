# Configuration

Provide a YAML file with `--config`, environment variables with the `GFW_` prefix, or Cobra flags. Priority is flags over environment over YAML over defaults.

```yaml
environment: production
listen_address: ":8080"
metrics_listen_address: ":9090"
database_url: "postgres://gfw:gfw@postgres:5432/gfw?sslmode=disable"
public_url: "https://gfw.example.com"
frontend_dir: "/app/frontend/.output/public"
oidc_issuer: "https://identity.example.com"
oidc_client_id: "great-firewall"
oidc_client_secret: "replace-me"
oidc_redirect_url: "https://gfw.example.com/auth/callback"
session_secret: "replace-with-at-least-32-bytes"
admin_token: "replace-with-at-least-32-random-bytes"
dev_auth: false
dns_interval: 15m
dns_timeout: 5s
retention_days: 365
```

`environment` accepts `development`, `staging`, or `production`. It identifies
the deployed instance in the interface and controls the color of the top status
line; it is not stored on machines or ecosystems.

The existing environment names remain supported, including `GFW_DATABASE_URL`, `GFW_DEV_AUTH`, `GFW_ADMIN_TOKEN`, and the OIDC variables documented in the README. Secrets should be supplied through the environment or a mounted secret rather than committed YAML.
