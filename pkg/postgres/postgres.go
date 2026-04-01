package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Config struct {
	Host               string
	Port               string
	User               string
	Password           string
	DBName             string
	SSLMode            string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetimeMin int
	Schema             string // primary schema (default: public)
	Schemas            string // optional comma-separated search_path schemas
}

func LoadConfigFromEnv() (Config, error) {
	get := func(key string) string { return os.Getenv(key) }

	maxOpen, err := atoiFallback(get("DB_MAX_OPEN_CONNS"), 10)
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}
	maxIdle, err := atoiFallback(get("DB_MAX_IDLE_CONNS"), 5)
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}
	lifetimeMin, err := atoiFallback(get("DB_CONN_MAX_LIFETIME_MIN"), 30)
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME_MIN: %w", err)
	}

	ssl := get("DB_SSLMODE")
	if ssl == "" {
		ssl = "disable"
	}

	cfg := Config{
		Host:               get("DB_HOST"),
		Port:               get("DB_PORT"),
		User:               get("DB_USER"),
		Password:           get("DB_PASSWORD"),
		DBName:             get("DB_NAME"),
		SSLMode:            ssl,
		MaxOpenConns:       maxOpen,
		MaxIdleConns:       maxIdle,
		ConnMaxLifetimeMin: lifetimeMin,
		Schema:             get("DB_SCHEMA"),
		Schemas:            get("DB_SCHEMAS"),
	}

	// минимальная валидация обязательных полей
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.DBName == "" {
		return Config{}, fmt.Errorf("missing required DB env vars (DB_HOST, DB_PORT, DB_USER, DB_NAME)")
	}

	return cfg, nil
}

// Connect открывает соединение и пингует Postgres с ретраями.
func Connect(cfg Config) (*sql.DB, error) {
	if db != nil {
		return db, nil
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// ВАЖНО: search_path должен быть в DSN, чтобы применяться ко всем соединениям из пула.
	searchPath := cfg.Schemas
	if searchPath == "" && cfg.Schema != "" {
		searchPath = cfg.Schema
	}
	if searchPath == "" {
		searchPath = "public"
	}
	if searchPath != "" {
		dsn += fmt.Sprintf(" search_path=%s", searchPath)
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// pool settings
	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMin) * time.Minute)

	// ping with retries (docker-friendly)
	const maxAttempts = 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := conn.Ping(); err == nil {
			log.Printf("[DB] connected to Postgres on attempt %d ✅", attempt)
			db = conn
			return db, nil
		} else {
			log.Printf("[DB] Postgres not ready yet (attempt %d/%d): %v", attempt, maxAttempts, err)
			time.Sleep(1 * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to ping postgres after retries")
}

func GetDB() *sql.DB { return db }

func Close() {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("[DB] error closing Postgres: %v\n", err)
		} else {
			log.Println("[DB] Postgres connection closed")
		}
	}
}

func atoiFallback(s string, def int) (int, error) {
	if s == "" {
		return def, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return v, nil
}
