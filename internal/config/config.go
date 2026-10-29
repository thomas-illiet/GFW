// Package config loads and validates Great Firewall runtime configuration.
package config

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config contains all runtime settings shared by the CLI and application.
type Config struct {
	Environment    string        `mapstructure:"environment" yaml:"environment"`
	ListenAddress  string        `mapstructure:"listen_address" yaml:"listen_address"`
	MetricsAddress string        `mapstructure:"metrics_listen_address" yaml:"metrics_listen_address"`
	DatabaseURL    string        `mapstructure:"database_url" yaml:"database_url"`
	PublicURL      string        `mapstructure:"public_url" yaml:"public_url"`
	FrontendDir    string        `mapstructure:"frontend_dir" yaml:"frontend_dir"`
	OIDCIssuer     string        `mapstructure:"oidc_issuer" yaml:"oidc_issuer"`
	OIDCClientID   string        `mapstructure:"oidc_client_id" yaml:"oidc_client_id"`
	OIDCSecret     string        `mapstructure:"oidc_client_secret" yaml:"oidc_client_secret"`
	OIDCRedirect   string        `mapstructure:"oidc_redirect_url" yaml:"oidc_redirect_url"`
	SessionSecret  string        `mapstructure:"session_secret" yaml:"session_secret"`
	AdminToken     string        `mapstructure:"admin_token" yaml:"admin_token"`
	DevAuth        bool          `mapstructure:"dev_auth" yaml:"dev_auth"`
	DNSInterval    time.Duration `mapstructure:"dns_interval" yaml:"dns_interval"`
	DNSTimeout     time.Duration `mapstructure:"dns_timeout" yaml:"dns_timeout"`
	RetentionDays  int           `mapstructure:"retention_days" yaml:"retention_days"`
}

// Loader owns an isolated Viper instance so configuration loading is testable.
type Loader struct{ v *viper.Viper }

// NewLoader returns a loader configured with defaults and GFW environment bindings.
func NewLoader() *Loader {
	v := viper.New()
	v.SetEnvPrefix("GFW")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	defaults := map[string]any{
		"environment":    "development",
		"listen_address": ":8080", "metrics_listen_address": ":9090", "database_url": "postgres://gfw:gfw@localhost:5432/gfw?sslmode=disable",
		"public_url": "http://localhost:8080", "frontend_dir": "frontend/.output/public",
		"oidc_redirect_url": "http://localhost:8080/auth/callback", "dev_auth": false,
		"dns_interval": 15 * time.Minute, "dns_timeout": 5 * time.Second, "retention_days": 365,
	}
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	return &Loader{v: v}
}

// BindFlags makes Cobra flags the highest-priority configuration source.
func (l *Loader) BindFlags(flags *pflag.FlagSet) error {
	bindings := map[string]string{
		"environment":    "environment",
		"listen_address": "listen-address", "metrics_listen_address": "metrics-listen-address", "database_url": "database-url", "public_url": "public-url",
		"frontend_dir": "frontend-dir", "oidc_issuer": "oidc-issuer", "oidc_client_id": "oidc-client-id",
		"oidc_client_secret": "oidc-client-secret", "oidc_redirect_url": "oidc-redirect-url",
		"session_secret": "session-secret", "admin_token": "admin-token", "dev_auth": "dev-auth", "dns_interval": "dns-interval",
		"dns_timeout": "dns-timeout", "retention_days": "retention-days",
	}
	for key, name := range bindings {
		if flag := flags.Lookup(name); flag != nil {
			if err := l.v.BindPFlag(key, flag); err != nil {
				return err
			}
		}
	}
	return nil
}

// Load merges defaults, an optional YAML file, environment variables, and flags.
func (l *Loader) Load(path string) (Config, error) {
	if path != "" {
		l.v.SetConfigFile(path)
		if err := l.v.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}
	var cfg Config
	if err := l.v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate rejects unsafe or incomplete runtime configuration.
func (c Config) Validate() error {
	switch c.Environment {
	case "development", "staging", "production":
	default:
		return errors.New("GFW_ENVIRONMENT must be one of development, staging, or production")
	}
	_, apiPort, apiErr := net.SplitHostPort(c.ListenAddress)
	_, metricsPort, metricsErr := net.SplitHostPort(c.MetricsAddress)
	if apiErr != nil || metricsErr != nil || apiPort == "" || metricsPort == "" {
		return errors.New("GFW_LISTEN_ADDRESS and GFW_METRICS_LISTEN_ADDRESS must be valid host:port addresses")
	}
	if apiPort == metricsPort {
		return errors.New("GFW_METRICS_LISTEN_ADDRESS must use a different port from GFW_LISTEN_ADDRESS")
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return errors.New("GFW_DATABASE_URL is required")
	}
	if c.RetentionDays < 1 {
		return errors.New("GFW_RETENTION_DAYS must be positive")
	}
	if c.DNSInterval <= 0 || c.DNSTimeout <= 0 {
		return errors.New("DNS intervals must be positive")
	}
	if !c.DevAuth && (c.OIDCIssuer == "" || c.OIDCClientID == "" || c.OIDCSecret == "" || len(c.SessionSecret) < 32) {
		return errors.New("OIDC settings and a 32-byte GFW_SESSION_SECRET are required unless GFW_DEV_AUTH=true")
	}
	if !c.DevAuth && len(c.AdminToken) < 32 {
		return errors.New("a 32-byte GFW_ADMIN_TOKEN is required unless GFW_DEV_AUTH=true")
	}
	return nil
}
