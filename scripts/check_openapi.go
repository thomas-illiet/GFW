// Command check-openapi validates the checked-in API document and route inventory.
package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type document struct {
	OpenAPI string                    `yaml:"openapi"`
	Info    struct{ Version string }  `yaml:"info"`
	Paths   map[string]map[string]any `yaml:"paths"`
}

var required = map[string][]string{
	"/health/live": {"get"}, "/health/ready": {"get"},
	"/api/v1/public/instance": {"get"},
	"/auth/login":             {"get"}, "/auth/callback": {"get"}, "/auth/logout": {"post"}, "/auth/session": {"get"},
	"/api/v1/firewall/ecosystems/{ecosystem}/ips": {"get"}, "/api/v1/firewall/ips/{ipv4}": {"get"},
	"/api/v1/admin/dashboard": {"get"}, "/api/v1/admin/ecosystems": {"get", "post"},
	"/api/v1/admin/ecosystems/{id}": {"patch", "delete"}, "/api/v1/admin/machines": {"get", "post"},
	"/api/v1/admin/machines/{id}": {"delete"}, "/api/v1/admin/trusted-services": {"get", "post"},
	"/api/v1/admin/trusted-services/{id}": {"patch", "delete"}, "/api/v1/admin/trusted-services/resolve": {"post"},
	"/api/v1/admin/ingestion-keys": {"get", "post"}, "/api/v1/admin/ingestion-keys/{id}": {"delete"},
	"/api/v1/ingestion/machines": {"post"},
}

func main() {
	body, err := os.ReadFile("openapi.yaml")
	if err != nil {
		fail(err)
	}
	var spec document
	if err := yaml.Unmarshal(body, &spec); err != nil {
		fail(err)
	}
	if spec.OpenAPI != "3.1.0" || spec.Info.Version != "1.0.0" {
		fail(fmt.Errorf("expected OpenAPI 3.1.0 contract version 1.0.0"))
	}
	for path, methods := range required {
		for _, method := range methods {
			if spec.Paths[path][method] == nil {
				fail(fmt.Errorf("missing operation %s %s", method, path))
			}
		}
	}
	if len(spec.Paths) != len(required) {
		fail(fmt.Errorf("route inventory mismatch: got %d paths, expected %d", len(spec.Paths), len(required)))
	}
}

func fail(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "openapi:", err)
	os.Exit(1)
}
