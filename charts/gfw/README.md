# GFW Helm chart

This chart deploys the complete GFW application tier: HTTP server, schema
migration hook, Service, optional Ingress and PodDisruptionBudget.
It intentionally contains no PostgreSQL workload or dependency; the runtime
uses an externally managed PostgreSQL endpoint.

All containers run as UID/GID `1000`, without privilege escalation or Linux
capabilities, with `RuntimeDefault` seccomp and a read-only root filesystem.
Only a bounded `emptyDir` is mounted at `/tmp`; product configuration is mounted
read-only at `/etc/gfw/config.yaml`. Kubernetes API credentials are not mounted.

The Service exposes the application on `service.port` and the isolated
Prometheus endpoint on `service.metricsPort` (9090 by default). Scrape
`/metrics` on the `metrics` port; this endpoint is not routed through the
application Ingress.

Create a secret (recommended):

```sh
kubectl -n gfw create secret generic gfw-runtime \
  --from-literal=database-url='postgres://...' \
  --from-literal=oidc-client-secret='...' \
  --from-literal=session-secret='at-least-32-random-bytes'
  --from-literal=admin-token='at-least-32-random-bytes'
```

Then install:

```sh
helm upgrade --install gfw ./charts/gfw --namespace gfw --create-namespace \
  --set secrets.existingSecret=gfw-runtime \
  --set config.environment=production \
  --set config.public_url=https://gfw.example.com \
  --set config.oidc_issuer=https://identity.example.com \
  --set config.oidc_client_id=great-firewall \
  --set config.oidc_redirect_url=https://gfw.example.com/auth/callback \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=gfw.example.com
```

Because schema migration is a `pre-install` hook, an `existingSecret` must
already exist before installation. When inline secret values are used instead,
the migration hook receives only `secrets.databaseURL`; the regular Secret is
then created with the release for the server.

For production, configure the Ingress TLS fields. This chart intentionally does
not create a NetworkPolicy: apply networking rules separately according to the
cluster CNI and the required DNS, OIDC and PostgreSQL destinations.

