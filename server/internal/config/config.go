package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr string
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

	return Config{Addr: ":" + strconv.Itoa(addr)}, nil
}
