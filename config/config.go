package config

import "os"

// Config contains database configuration shared by both scraper and web projects.
type Config struct {
	DatabaseURL string
}

// Default returns configuration values, reading from environment variables first
// and falling back to development defaults.
func Default() Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://snowfinder:snowfinder@localhost:5432/snowfinder?sslmode=disable"
	}
	return Config{
		DatabaseURL: dbURL,
	}
}
