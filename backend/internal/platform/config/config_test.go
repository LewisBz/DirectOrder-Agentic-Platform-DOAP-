package config

import (
	"strings"
	"testing"
)

func TestLoadFailsNamingMissingKeys(t *testing.T) {
	t.Setenv("API_ADDR", "")
	t.Setenv("POSTGRES_HOST", "")
	t.Setenv("POSTGRES_PORT", "")
	t.Setenv("POSTGRES_DB", "")
	t.Setenv("POSTGRES_APP_USER", "")
	t.Setenv("POSTGRES_APP_PASSWORD", "")
	t.Setenv("AUTH_JWT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	for _, key := range []string{"API_ADDR", "POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB", "POSTGRES_APP_USER", "POSTGRES_APP_PASSWORD", "AUTH_JWT_SECRET"} {
		if !strings.Contains(msg, key) {
			t.Fatalf("error %q should name %s", msg, key)
		}
	}
}

func TestLoadOK(t *testing.T) {
	t.Setenv("API_ADDR", ":8080")
	t.Setenv("POSTGRES_HOST", "postgres")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "doap")
	t.Setenv("POSTGRES_APP_USER", "doap_app")
	t.Setenv("POSTGRES_APP_PASSWORD", "secret")
	t.Setenv("AUTH_JWT_SECRET", "changeme_jwt_secret_do_not_use_prod_xx")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIAddr != ":8080" {
		t.Fatalf("addr %s", cfg.APIAddr)
	}
}

func TestLoadRejectsShortJWTSecret(t *testing.T) {
	t.Setenv("API_ADDR", ":8080")
	t.Setenv("POSTGRES_HOST", "postgres")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "doap")
	t.Setenv("POSTGRES_APP_USER", "doap_app")
	t.Setenv("POSTGRES_APP_PASSWORD", "secret")
	t.Setenv("AUTH_JWT_SECRET", "too-short")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "AUTH_JWT_SECRET") {
		t.Fatalf("error %q", err)
	}
}
