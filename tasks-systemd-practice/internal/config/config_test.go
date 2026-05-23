package config

import "testing"

func TestFromEnvUsesDefaults(t *testing.T) {
	t.Setenv("TASKS_HOST", "")
	t.Setenv("TASKS_PORT", "")
	t.Setenv("LOG_LEVEL", "")

	cfg := FromEnv()

	if cfg.Host != defaultHost {
		t.Fatalf("expected default host %q, got %q", defaultHost, cfg.Host)
	}

	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %q, got %q", defaultPort, cfg.Port)
	}

	if cfg.LogLevel != defaultLogLevel {
		t.Fatalf("expected default log level %q, got %q", defaultLogLevel, cfg.LogLevel)
	}
}

func TestFromEnvNormalizesPort(t *testing.T) {
	t.Setenv("TASKS_PORT", ":9090")

	cfg := FromEnv()

	if cfg.Port != "9090" {
		t.Fatalf("expected normalized port 9090, got %q", cfg.Port)
	}
}
