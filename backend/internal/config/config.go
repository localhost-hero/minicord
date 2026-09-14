package config

import (
	"fmt"
	"strings"
)

type Config struct {
	BackendAddr       string
	LiveKitURL        string
	LiveKitAPIKey     string
	LiveKitAPISecret  string
	DBPath            string
	AdminUsername     string
	AdminPasswordHash string
	SessionSecret     string
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

	dbPath := getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/minicord.db"
	}

	adminUsername := strings.TrimSpace(getenv("ADMIN_USERNAME"))
	if adminUsername == "" {
		return Config{}, fmt.Errorf("ADMIN_USERNAME must not be empty")
	}

	adminPasswordHash := strings.TrimSpace(getenv("ADMIN_PASSWORD_HASH"))
	if adminPasswordHash == "" {
		return Config{}, fmt.Errorf("ADMIN_PASSWORD_HASH must not be empty")
	}

	sessionSecret := getenv("SESSION_SECRET")
	if len(sessionSecret) < 32 {
		return Config{}, fmt.Errorf("SESSION_SECRET must be at least 32 characters")
	}

	return Config{
		BackendAddr:       address,
		LiveKitURL:        liveKitURL,
		LiveKitAPIKey:     apiKey,
		LiveKitAPISecret:  apiSecret,
		DBPath:            dbPath,
		AdminUsername:     adminUsername,
		AdminPasswordHash: adminPasswordHash,
		SessionSecret:     sessionSecret,
	}, nil
}

