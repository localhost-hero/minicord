package config

import (
	"fmt"
	"strings"
)

type Config struct {
	BackendAddr string
}

func Load(getenv func(string) string) (Config, error) {
	address := getenv("BACKEND_ADDR")
	if address == "" {
		address = ":8080"
	}

	if strings.TrimSpace(address) == "" {
		return Config{}, fmt.Errorf("BACKEND_ADDR must not be empty")
	}

	return Config{BackendAddr: address}, nil
}
