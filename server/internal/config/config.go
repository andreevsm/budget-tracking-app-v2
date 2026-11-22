package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr       string
	DSN        string
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func Load() (Config, error) {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	addr, err := strconv.Atoi(port)

	if err != nil {
		return Config{}, fmt.Errorf("parsing port: %q, %w", port, err)
	}

	if addr < 1 || addr > 65535 {
		return Config{}, fmt.Errorf("port should be between 1 and 65535")
	}

	dsn, ok := os.LookupEnv("DB_DSN")

	if !ok || dsn == "" {
		return Config{}, fmt.Errorf("DB_DSN is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	return Config{Addr: ":" + strconv.Itoa(addr), DSN: dsn}, nil
}
