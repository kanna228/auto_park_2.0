package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"auto_park/pkg/postgres"
)

type Config struct {
	Service ServiceConfig
	DB      postgres.Config
	Schemas SchemaConfig
	Auth    AuthConfig
	Email   EmailConfig
}

type ServiceConfig struct {
	Name string
	Port string
}

type SchemaConfig struct {
	User      string
	Vehicle   string
	Tripsheet string
}

type AuthConfig struct {
	JWTSecret      string
	TokenTTL       time.Duration
	CookieDomain   string
	CookieSecure   bool
	CookieSameSite string
}

type EmailConfig struct {
	SendEmail bool

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	FromName string
	FromAddr string
}

func Load() (*Config, error) {
	dbCfg, err := postgres.LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Service: ServiceConfig{
			Name: getenv("SERVICE_NAME", "auto_park"),
			Port: getenv("SERVICE_PORT", "8080"),
		},
		DB: dbCfg,
		Schemas: SchemaConfig{
			User:      getenv("USER_DB_SCHEMA", "public"),
			Vehicle:   getenv("VEHICLE_DB_SCHEMA", "public"),
			Tripsheet: getenv("TRIPSHEET_DB_SCHEMA", "public"),
		},
		Auth: AuthConfig{
			JWTSecret:      getenv("JWT_SECRET", "dev_secret_change_me"),
			TokenTTL:       time.Duration(getenvInt("JWT_TTL_HOURS", 24)) * time.Hour,
			CookieDomain:   getenv("COOKIE_DOMAIN", ""),
			CookieSecure:   getenvBool("COOKIE_SECURE", false),
			CookieSameSite: getenv("COOKIE_SAMESITE", "Lax"),
		},
		Email: EmailConfig{
			SendEmail:    getenvBool("SEND_EMAIL", false),
			SMTPHost:     getenv("SMTP_HOST", ""),
			SMTPPort:     getenvInt("SMTP_PORT", 587),
			SMTPUser:     getenv("SMTP_USER", ""),
			SMTPPassword: getenv("SMTP_PASSWORD", ""),
			FromName:     getenv("SMTP_FROM_NAME", "Autopark"),
			FromAddr:     getenv("SMTP_FROM_ADDR", ""),
		},
	}

	return cfg, nil
}

func getenv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func getenvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func getenvBool(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}
