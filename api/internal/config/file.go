package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// fileConfig mirrors the TOML layout. Omitted keys keep their prior value during merge.
type fileConfig struct {
	Server     serverFileConfig     `toml:"server"`
	Database   databaseFileConfig   `toml:"database"`
	Valkey     valkeyFileConfig     `toml:"valkey"`
	Meilisearch meilisearchFileConfig `toml:"meilisearch"`
	S3         s3FileConfig         `toml:"s3"`
	Email      emailFileConfig      `toml:"email"`
	SMTP       smtpFileConfig       `toml:"smtp"`
	Mailgun    mailgunFileConfig    `toml:"mailgun"`
	Discord    discordFileConfig    `toml:"discord"`
	WebAuthn   webauthnFileConfig   `toml:"webauthn"`
	Anthropic  anthropicFileConfig  `toml:"anthropic"`
	OVHAI      ovhAIFileConfig      `toml:"ovh_ai"`
	OpenFoodFacts openFoodFactsFileConfig `toml:"openfoodfacts"`
}

type serverFileConfig struct {
	Port      string `toml:"port"`
	BaseURL   string `toml:"base_url"`
	WebOrigin string `toml:"web_origin"`
	JWTSecret string `toml:"jwt_secret"`
}

type databaseFileConfig struct {
	URL string `toml:"url"`
}

type valkeyFileConfig struct {
	URL string `toml:"url"`
}

type meilisearchFileConfig struct {
	URL    string `toml:"url"`
	APIKey string `toml:"api_key"`
}

type s3FileConfig struct {
	Endpoint              string `toml:"endpoint"`
	Region                string `toml:"region"`
	AccessKey             string `toml:"access_key"`
	SecretKey             string `toml:"secret_key"`
	ClientAssetsBucket    string `toml:"client_assets_bucket"`
	StaticAssetsBucket    string `toml:"static_assets_bucket"`
	UsePathStyle          *bool  `toml:"use_path_style"`
	PublicBaseURL         string `toml:"public_base_url"`
	PresignPrivateObjects *bool  `toml:"presign_private_objects"`
	PresignExpirySeconds  *int   `toml:"presign_expiry_seconds"`
}

type emailFileConfig struct {
	Provider string `toml:"provider"`
	From     string `toml:"from"`
}

type smtpFileConfig struct {
	Host string `toml:"host"`
	Port *int   `toml:"port"`
	From string `toml:"from"`
}

type mailgunFileConfig struct {
	APIKey     string `toml:"api_key"`
	Domain     string `toml:"domain"`
	APIBaseURL string `toml:"api_base_url"`
}

type discordFileConfig struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	RedirectURI  string `toml:"redirect_uri"`
}

type webauthnFileConfig struct {
	RPID   string `toml:"rp_id"`
	Origin string `toml:"rp_origin"`
	Name   string `toml:"rp_name"`
}

type anthropicFileConfig struct {
	APIKey string `toml:"api_key"`
	Model  string `toml:"model"`
}

type ovhAIFileConfig struct {
	AccessToken string `toml:"access_token"`
	Model       string `toml:"model"`
	BaseURL     string `toml:"base_url"`
}

type openFoodFactsFileConfig struct {
	SearchURL string `toml:"search_url"`
	BaseURL   string `toml:"base_url"`
}

func loadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}

	var file fileConfig
	if err := toml.Unmarshal(data, &file); err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}

	cfg := Config{}
	applyFile(&cfg, file)
	return cfg, nil
}

func applyFile(cfg *Config, file fileConfig) {
	setString(&cfg.Port, file.Server.Port)
	setString(&cfg.BaseURL, file.Server.BaseURL)
	setString(&cfg.WebOrigin, file.Server.WebOrigin)
	setString(&cfg.JWTSecret, file.Server.JWTSecret)

	setString(&cfg.DatabaseURL, file.Database.URL)
	setString(&cfg.ValkeyURL, file.Valkey.URL)

	setString(&cfg.MeilisearchURL, file.Meilisearch.URL)
	setString(&cfg.MeilisearchAPIKey, file.Meilisearch.APIKey)

	setString(&cfg.S3Endpoint, file.S3.Endpoint)
	setString(&cfg.S3Region, file.S3.Region)
	setString(&cfg.S3AccessKey, file.S3.AccessKey)
	setString(&cfg.S3SecretKey, file.S3.SecretKey)
	setString(&cfg.S3ClientBucket, file.S3.ClientAssetsBucket)
	setString(&cfg.S3StaticBucket, file.S3.StaticAssetsBucket)
	if file.S3.UsePathStyle != nil {
		cfg.S3UsePathStyle = *file.S3.UsePathStyle
		cfg.s3UsePathStyleSet = true
	}
	setString(&cfg.S3PublicBaseURL, file.S3.PublicBaseURL)
	if file.S3.PresignPrivateObjects != nil {
		cfg.S3PresignPrivateObjects = *file.S3.PresignPrivateObjects
		cfg.s3PresignPrivateObjectsSet = true
	}
	if file.S3.PresignExpirySeconds != nil {
		cfg.S3PresignExpirySeconds = *file.S3.PresignExpirySeconds
		cfg.s3PresignExpirySecondsSet = true
	}

	setString(&cfg.EmailProvider, file.Email.Provider)
	setString(&cfg.EmailFrom, file.Email.From)
	setString(&cfg.SMTPHost, file.SMTP.Host)
	if file.SMTP.Port != nil {
		cfg.SMTPPort = *file.SMTP.Port
		cfg.smtpPortSet = true
	}
	if strings.TrimSpace(cfg.EmailFrom) == "" {
		setString(&cfg.EmailFrom, file.SMTP.From)
	}
	setString(&cfg.MailgunAPIKey, file.Mailgun.APIKey)
	setString(&cfg.MailgunDomain, file.Mailgun.Domain)
	setString(&cfg.MailgunAPIBaseURL, file.Mailgun.APIBaseURL)

	setString(&cfg.DiscordClientID, file.Discord.ClientID)
	setString(&cfg.DiscordSecret, file.Discord.ClientSecret)
	setString(&cfg.DiscordRedirect, file.Discord.RedirectURI)

	setString(&cfg.WebAuthnRPID, file.WebAuthn.RPID)
	setString(&cfg.WebAuthnOrigin, file.WebAuthn.Origin)
	setString(&cfg.WebAuthnRPName, file.WebAuthn.Name)

	setString(&cfg.AnthropicAPIKey, file.Anthropic.APIKey)
	setString(&cfg.AnthropicModel, file.Anthropic.Model)
	setString(&cfg.OVHAIToken, file.OVHAI.AccessToken)
	setString(&cfg.OVHAIModel, file.OVHAI.Model)
	setString(&cfg.OVHAIBaseURL, file.OVHAI.BaseURL)
	setString(&cfg.OFFSearchURL, file.OpenFoodFacts.SearchURL)
	setString(&cfg.OFFBaseURL, file.OpenFoodFacts.BaseURL)
}

func setString(dst *string, value string) {
	if strings.TrimSpace(value) != "" {
		*dst = strings.TrimSpace(value)
	}
}
