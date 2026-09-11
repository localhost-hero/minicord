package config

import "testing"

func TestLoadUsesDefaultBackendAddress(t *testing.T) {
	configuration, err := Load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if configuration.BackendAddr != ":8080" {
		t.Fatalf("expected default address %q, got %q", ":8080", configuration.BackendAddr)
	}
}

func TestLoadRejectsWhitespaceOnlyBackendAddress(t *testing.T) {
	_, err := Load(func(string) string { return "   " })
	if err == nil {
		t.Fatal("expected an empty address error")
	}
}
