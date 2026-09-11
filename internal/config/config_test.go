package config

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndBuildDatabaseURL(t *testing.T) {
	path := writeConfig(t, `
server:
  port: 9090
database:
  host: 127.0.0.1
  port: 5432
  user: postgres
  password: "p@ss:word"
  name: blog
  ssl_mode: disable
`)

	configuration, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if configuration.Address() != ":9090" {
		t.Fatalf("Address() = %q", configuration.Address())
	}

	databaseURL, err := url.Parse(configuration.DatabaseURL())
	if err != nil {
		t.Fatalf("DatabaseURL() parse error = %v", err)
	}
	password, ok := databaseURL.User.Password()
	if !ok || password != "p@ss:word" {
		t.Fatalf("DatabaseURL() did not preserve password")
	}
	if databaseURL.Host != "127.0.0.1:5432" || databaseURL.Path != "/blog" {
		t.Fatalf("DatabaseURL() = %q", databaseURL.String())
	}
}

func TestLoadRejectsMissingDatabaseName(t *testing.T) {
	path := writeConfig(t, `
database:
  user: postgres
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
