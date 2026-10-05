package redis

import (
	"context"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func resetClient() {
	mu.Lock()
	client = nil
	mu.Unlock()
}

func TestFromEnvAndEnabled(t *testing.T) {
	t.Setenv("AIRWAY_REDIS", "")
	t.Setenv("REDIS", "")

	if FromEnv().Enabled() {
		t.Fatalf("expected disabled without env")
	}

	t.Setenv("REDIS", "redis://localhost:6379/0")
	if cfg := FromEnv(); !cfg.Enabled() || cfg.URL != "redis://localhost:6379/0" {
		t.Fatalf("REDIS not picked up: %+v", cfg)
	}

	t.Setenv("AIRWAY_REDIS", "redis://other:6379/1")
	if cfg := FromEnv(); cfg.URL != "redis://other:6379/1" {
		t.Fatalf("AIRWAY_REDIS should win over REDIS: %+v", cfg)
	}
}

func TestSetupWithoutURL(t *testing.T) {
	if err := Setup(Config{}); err == nil {
		t.Fatalf("expected error for empty URL")
	}
}

func TestSetupRejectsBadURL(t *testing.T) {
	if err := Setup(Config{URL: "://nope"}); err == nil {
		t.Fatalf("expected error for invalid URL")
	}
}

func TestSetupRejectsUnreachable(t *testing.T) {
	if err := Setup(Config{URL: "redis://127.0.0.1:1/0"}); err == nil ||
		!strings.Contains(err.Error(), "ping failed") {
		t.Fatalf("expected ping failure, got %v", err)
	}
}

func TestSetupPingCurrentAgainstMiniredis(t *testing.T) {
	server := miniredis.RunT(t)

	resetClient()
	if Current() != nil {
		t.Fatalf("expected nil client before Setup")
	}

	if err := Setup(Config{URL: "redis://" + server.Addr()}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if Current() == nil || Must() != Current() {
		t.Fatalf("expected a live client after Setup")
	}
	if err := Ping(context.Background()); err != nil {
		t.Fatalf("ping: %v", err)
	}

	// The client is a real go-redis handle.
	if err := Must().Set(context.Background(), "k", "v", 0).Err(); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := server.Get("k")
	if err != nil || got != "v" {
		t.Fatalf("expected miniredis to hold k=v, got %q (%v)", got, err)
	}

	server.Close()
	if err := Ping(context.Background()); err == nil {
		t.Fatalf("expected ping error after server shutdown")
	}

	resetClient()
}

func TestMustPanicsWithoutSetup(t *testing.T) {
	resetClient()
	defer resetClient()

	defer func() {
		if recover() == nil {
			t.Fatalf("expected Must to panic without Setup")
		}
	}()
	Must()
}
