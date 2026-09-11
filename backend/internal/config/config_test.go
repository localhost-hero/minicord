package config

import "testing"

func validEnv(overrides map[string]string) func(string) string {
	values := map[string]string{
		"LIVEKIT_URL":        "ws://localhost:7880",
		"LIVEKIT_API_KEY":    "test-key",
		"LIVEKIT_API_SECRET": "test-secret",
	}
	for key, value := range overrides {
		values[key] = value
	}
	return func(key string) string { return values[key] }
}

func TestLoadUsesDefaultBackendAddress(t *testing.T) {
	configuration, err := Load(validEnv(nil))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if configuration.BackendAddr != ":8080" {
		t.Fatalf("expected default address %q, got %q", ":8080", configuration.BackendAddr)
	}
}

func TestLoadRejectsWhitespaceOnlyBackendAddress(t *testing.T) {
	_, err := Load(validEnv(map[string]string{"BACKEND_ADDR": "   "}))
	if err == nil {
		t.Fatal("expected an empty address error")
	}
}

func TestLoadRejectsMissingLiveKitURL(t *testing.T) {
	_, err := Load(validEnv(map[string]string{"LIVEKIT_URL": ""}))
	if err == nil {
		t.Fatal("expected a missing LIVEKIT_URL error")
	}
}

func TestLoadRejectsMissingLiveKitAPIKey(t *testing.T) {
	_, err := Load(validEnv(map[string]string{"LIVEKIT_API_KEY": ""}))
	if err == nil {
		t.Fatal("expected a missing LIVEKIT_API_KEY error")
	}
}

func TestLoadRejectsMissingLiveKitAPISecret(t *testing.T) {
	_, err := Load(validEnv(map[string]string{"LIVEKIT_API_SECRET": ""}))
	if err == nil {
		t.Fatal("expected a missing LIVEKIT_API_SECRET error")
	}
}

func TestLoadReturnsLiveKitConfiguration(t *testing.T) {
	configuration, err := Load(validEnv(nil))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if configuration.LiveKitURL != "ws://localhost:7880" {
		t.Fatalf("unexpected LiveKitURL: %q", configuration.LiveKitURL)
	}
	if configuration.LiveKitAPIKey != "test-key" {
		t.Fatalf("unexpected LiveKitAPIKey: %q", configuration.LiveKitAPIKey)
	}
	if configuration.LiveKitAPISecret != "test-secret" {
		t.Fatalf("unexpected LiveKitAPISecret: %q", configuration.LiveKitAPISecret)
	}
}
