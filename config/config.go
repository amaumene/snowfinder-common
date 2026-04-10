package config

import "os"

// Config contains database configuration shared by both scraper and web projects.
type Config struct {
	DatabasePath string
}

// Default reads configuration values from environment variables.
// DATABASE_PATH must be set; callers should validate before connecting.
func Default() Config {
	return Config{
		DatabasePath: os.Getenv("DATABASE_PATH"),
	}
}
