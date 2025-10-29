package config

// Config contains database configuration shared by both scraper and web projects.
type Config struct {
	DatabaseURL string
}

// Default returns default configuration values.
func Default() Config {
	return Config{
		DatabaseURL: "postgres://snowfinder:snowfinder@localhost:5432/snowfinder?sslmode=disable",
	}
}
