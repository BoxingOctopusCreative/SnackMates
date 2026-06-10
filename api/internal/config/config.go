package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	BaseURL           string
	WebOrigin         string
	JWTSecret         string
	DatabaseURL       string
	ValkeyURL         string
	MeilisearchURL    string
	MeilisearchAPIKey string
	S3Endpoint        string
	S3Region          string
	S3AccessKey       string
	S3SecretKey       string
	S3ClientBucket           string
	S3StaticBucket           string
	S3UsePathStyle           bool
	S3PublicBaseURL          string
	S3PresignPrivateObjects  bool
	S3PresignExpirySeconds   int
	EmailProvider     string
	EmailFrom         string
	SMTPHost          string
	SMTPPort          int
	MailgunAPIKey     string
	MailgunDomain     string
	MailgunAPIBaseURL string
	DiscordClientID   string
	DiscordSecret     string
	DiscordRedirect   string
	WebAuthnRPID      string
	WebAuthnOrigin    string
	WebAuthnRPName    string
	AnthropicAPIKey   string
	AnthropicModel    string
	OVHAIToken        string
	OVHAIModel        string
	OVHAIBaseURL      string
	OFFSearchURL      string
	OFFBaseURL        string
	SentryDSN         string
	SentryEnvironment string

	s3UsePathStyleSet          bool
	s3PresignPrivateObjectsSet bool
	s3PresignExpirySecondsSet  bool
	smtpPortSet                bool
}

// Options controls optional config file loading. When ConfigFile is empty, only
// environment variables (with built-in defaults) are used.
type Options struct {
	ConfigFile string
}

// Load resolves configuration from an optional TOML file and environment variables.
// Environment variables always override values from the config file. When no config
// file is specified, behavior matches env-only loading with the same defaults.
func Load(opts Options) (Config, error) {
	var cfg Config

	if path := strings.TrimSpace(opts.ConfigFile); path != "" {
		fileCfg, err := loadFile(path)
		if err != nil {
			return Config{}, err
		}
		cfg = fileCfg
	}

	applyEnv(&cfg)
	return cfg, nil
}

func applyEnv(cfg *Config) {
	cfg.Port = envString(cfg.Port, "API_PORT", "8080")
	cfg.BaseURL = envString(cfg.BaseURL, "API_BASE_URL", "http://localhost:8080")
	cfg.WebOrigin = envString(cfg.WebOrigin, "WEB_ORIGIN", "http://localhost:3000")
	cfg.JWTSecret = envString(cfg.JWTSecret, "JWT_SECRET", "dev-secret-change-me")

	cfg.DatabaseURL = envString(cfg.DatabaseURL, "DATABASE_URL", "postgres://snackmates:snackmates@localhost:5432/snackmates?sslmode=disable")
	cfg.ValkeyURL = envString(cfg.ValkeyURL, "VALKEY_URL", "redis://localhost:6379/0")

	cfg.MeilisearchURL = envString(cfg.MeilisearchURL, "MEILISEARCH_URL", "http://localhost:7700")
	cfg.MeilisearchAPIKey = envString(cfg.MeilisearchAPIKey, "MEILISEARCH_API_KEY", "snackmates-dev-master-key")

	cfg.S3Endpoint = envString(cfg.S3Endpoint, "S3_ENDPOINT", "http://localhost:9000")
	cfg.S3Region = envString(cfg.S3Region, "S3_REGION", "us-east-1")
	cfg.S3AccessKey = envString(cfg.S3AccessKey, "S3_ACCESS_KEY", "snackmates")
	cfg.S3SecretKey = envString(cfg.S3SecretKey, "S3_SECRET_KEY", "snackmates-dev")
	cfg.S3ClientBucket = envString(cfg.S3ClientBucket, "S3_CLIENT_ASSETS_BUCKET", "client-assets")
	cfg.S3StaticBucket = envString(cfg.S3StaticBucket, "S3_STATIC_ASSETS_BUCKET", "static-assets")
	cfg.S3UsePathStyle = envBool(cfg.S3UsePathStyle, cfg.s3UsePathStyleSet, "S3_USE_PATH_STYLE", true)
	cfg.S3PublicBaseURL = envString(cfg.S3PublicBaseURL, "S3_PUBLIC_BASE_URL", "")
	cfg.S3PresignPrivateObjects = envBool(cfg.S3PresignPrivateObjects, cfg.s3PresignPrivateObjectsSet, "S3_PRESIGN_PRIVATE_OBJECTS", true)
	cfg.S3PresignExpirySeconds = envInt(cfg.S3PresignExpirySeconds, cfg.s3PresignExpirySecondsSet, "S3_PRESIGN_EXPIRY_SECONDS", 3600)

	cfg.EmailProvider = envString(cfg.EmailProvider, "EMAIL_PROVIDER", "smtp")
	cfg.EmailFrom = envString(cfg.EmailFrom, "EMAIL_FROM", "")
	if strings.TrimSpace(cfg.EmailFrom) == "" {
		cfg.EmailFrom = envString(cfg.EmailFrom, "SMTP_FROM", "noreply@snackmates.local")
	}
	cfg.SMTPHost = envString(cfg.SMTPHost, "SMTP_HOST", "localhost")
	cfg.SMTPPort = envInt(cfg.SMTPPort, cfg.smtpPortSet, "SMTP_PORT", 1025)
	cfg.MailgunAPIKey = envString(cfg.MailgunAPIKey, "MAILGUN_API_KEY", "")
	cfg.MailgunDomain = envString(cfg.MailgunDomain, "MAILGUN_DOMAIN", "")
	cfg.MailgunAPIBaseURL = envString(cfg.MailgunAPIBaseURL, "MAILGUN_API_BASE_URL", "")

	cfg.DiscordClientID = envString(cfg.DiscordClientID, "DISCORD_CLIENT_ID", "")
	cfg.DiscordSecret = envString(cfg.DiscordSecret, "DISCORD_CLIENT_SECRET", "")
	cfg.DiscordRedirect = envString(cfg.DiscordRedirect, "DISCORD_REDIRECT_URI", "http://localhost:8080/api/v1/auth/discord/callback")

	cfg.WebAuthnRPID = envString(cfg.WebAuthnRPID, "WEBAUTHN_RP_ID", "localhost")
	cfg.WebAuthnOrigin = envString(cfg.WebAuthnOrigin, "WEBAUTHN_RP_ORIGIN", "http://localhost:3000")
	cfg.WebAuthnRPName = envString(cfg.WebAuthnRPName, "WEBAUTHN_RP_NAME", "SnackMates")

	cfg.AnthropicAPIKey = envString(cfg.AnthropicAPIKey, "ANTHROPIC_API_KEY", "")
	cfg.AnthropicModel = envString(cfg.AnthropicModel, "ANTHROPIC_MODEL", "claude-haiku-4-5-20251001")
	cfg.OVHAIToken = envString(cfg.OVHAIToken, "OVH_AI_ENDPOINTS_ACCESS_TOKEN", "")
	cfg.OVHAIModel = envString(cfg.OVHAIModel, "OVH_AI_MODEL", "Mistral-Nemo-Instruct-2407")
	cfg.OVHAIBaseURL = envString(cfg.OVHAIBaseURL, "OVH_AI_BASE_URL", "https://oai.endpoints.kepler.ai.cloud.ovh.net/v1")
	cfg.OFFSearchURL = envString(cfg.OFFSearchURL, "OPENFOODFACTS_SEARCH_URL", "https://search.openfoodfacts.org")
	cfg.OFFBaseURL = envString(cfg.OFFBaseURL, "OPENFOODFACTS_BASE_URL", "https://ca.openfoodfacts.org")

	cfg.SentryDSN = envString(cfg.SentryDSN, "SENTRY_DSN", "")
	cfg.SentryEnvironment = envString(cfg.SentryEnvironment, "SENTRY_ENVIRONMENT", "")
}

func envString(current, key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(v)
	}
	if strings.TrimSpace(current) != "" {
		return current
	}
	return fallback
}

func envBool(current bool, currentSet bool, key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		return strings.EqualFold(strings.TrimSpace(v), "true") || v == "1"
	}
	if currentSet {
		return current
	}
	return fallback
}

func envInt(current int, currentSet bool, key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return fallback
		}
		return n
	}
	if currentSet {
		return current
	}
	return fallback
}

// ResolveConfigPath returns the config file path from CLI flag or CONFIG_FILE env.
func ResolveConfigPath(flagValue string) string {
	if path := strings.TrimSpace(flagValue); path != "" {
		return path
	}
	return strings.TrimSpace(os.Getenv("CONFIG_FILE"))
}

// Validate checks required production-sensitive settings.
func (c Config) Validate() error {
	if strings.TrimSpace(c.JWTSecret) == "" {
		return fmt.Errorf("jwt secret must not be empty")
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("database url must not be empty")
	}
	return nil
}
