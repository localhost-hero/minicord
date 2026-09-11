package config

import (
	"fmt"
	"strings"
)

type Config struct {
	BackendAddr      string
	LiveKitURL       string
	LiveKitAPIKey    string
	LiveKitAPISecret string
}

func Load(getenv func(string) string) (Config, error) {
	address := getenv("BACKEND_ADDR")
	if address == "" {
		address = ":8080"
	}

	if strings.TrimSpace(address) == "" {
		return Config{}, fmt.Errorf("BACKEND_ADDR must not be empty")
	}

	liveKitURL := strings.TrimSpace(getenv("LIVEKIT_URL"))
	if liveKitURL == "" {
		return Config{}, fmt.Errorf("LIVEKIT_URL must not be empty")
	}

	apiKey := strings.TrimSpace(getenv("LIVEKIT_API_KEY"))
	if apiKey == "" {
		return Config{}, fmt.Errorf("LIVEKIT_API_KEY must not be empty")
	}

	apiSecret := strings.TrimSpace(getenv("LIVEKIT_API_SECRET"))
	if apiSecret == "" {
		return Config{}, fmt.Errorf("LIVEKIT_API_SECRET must not be empty")
	}

	return Config{
		BackendAddr:      address,
		LiveKitURL:       liveKitURL,
		LiveKitAPIKey:    apiKey,
		LiveKitAPISecret: apiSecret,
	}, nil
}
