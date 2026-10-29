package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

// TestLoadPrecedence verifies defaults, YAML, environment, and flag precedence.
func TestLoadPrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gfw.yaml")
	if err := os.WriteFile(path, []byte("listen_address: ':9000'\ndev_auth: true\nretention_days: 20\ndns_interval: 2m\ndns_timeout: 3s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GFW_RETENTION_DAYS", "30")
	loader := NewLoader()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("listen-address", "", "")
	flags.Int("retention-days", 0, "")
	if err := flags.Set("listen-address", ":9100"); err != nil {
		t.Fatal(err)
	}
	if err := loader.BindFlags(flags); err != nil {
		t.Fatal(err)
	}
	cfg, err := loader.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddress != ":9100" || cfg.RetentionDays != 30 || cfg.DNSInterval != 2*time.Minute {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

// TestLoadErrors covers file, decoding, and validation failures.
func TestLoadErrors(t *testing.T) {
	if _, err := NewLoader().Load("/does/not/exist"); err == nil {
		t.Fatal("expected file error")
	}
	t.Setenv("GFW_DATABASE_URL", " ")
	if _, err := NewLoader().Load(""); err == nil {
		t.Fatal("expected database validation error")
	}
	t.Setenv("GFW_DATABASE_URL", "postgres://db")
	t.Setenv("GFW_DEV_AUTH", "true")
	t.Setenv("GFW_RETENTION_DAYS", "0")
	if _, err := NewLoader().Load(""); err == nil {
		t.Fatal("expected retention error")
	}
	t.Setenv("GFW_RETENTION_DAYS", "1")
	t.Setenv("GFW_DNS_TIMEOUT", "0s")
	if _, err := NewLoader().Load(""); err == nil {
		t.Fatal("expected DNS error")
	}
}

// TestProductionValidation verifies required OIDC settings and the valid path.
func TestProductionValidation(t *testing.T) {
	cfg := Config{DatabaseURL: "postgres://db", RetentionDays: 1, DNSInterval: time.Minute, DNSTimeout: time.Second}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected OIDC error")
	}
	cfg.OIDCIssuer, cfg.OIDCClientID, cfg.OIDCSecret, cfg.SessionSecret = "issuer", "id", "secret", "01234567890123456789012345678901"
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}
