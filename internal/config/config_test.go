package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadDefaultsWhenMissing(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Host != DefaultHost || cfg.Port != DefaultPort {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	want := Config{
		Host:           "10.0.0.2",
		Port:           9090,
		NetworkService: "USB 10/100/1000 LAN",
		ShellType:      "zsh",
		NoProxyCustom:  []string{"internal.example.com", "*.corp.local"},
	}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("config file is empty")
	}
}

func TestLoadListSyntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `host: 127.0.0.1
port: 7897
shell_type: zsh
no_proxy_custom:
  - a.internal
  - "*.corp.local"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.NoProxyCustom) != 2 || got.NoProxyCustom[0] != "a.internal" || got.NoProxyCustom[1] != "*.corp.local" {
		t.Fatalf("unexpected no_proxy_custom: %+v", got.NoProxyCustom)
	}
}
