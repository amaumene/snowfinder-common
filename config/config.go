package config

import "os"

// Config contains database configuration shared by both scraper and web projects.
type Config struct {
	DatabaseURL string
}

// Default returns configuration values from environment variables.
// DATABASE_URL must be set; callers should validate before connecting.
func Default() Config {
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
