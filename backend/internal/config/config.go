package config

import "os"

type Config struct {
	Port           string
	DatabaseURL    string
	AnthropicAPIKey string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/shift_app?sslmode=disable"
	}

	return &Config{
		Port:           port,
		DatabaseURL:    dbURL,
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
}
