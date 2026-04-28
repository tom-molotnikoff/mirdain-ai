package config_test

import (
"os"
"testing"

"github.com/tom-molotnikoff/mirdain-ai/internal/config"
)

func TestLoad_MissingFile(t *testing.T) {
_, err := config.Load("/tmp/nonexistent_mirdain.yaml")
if err == nil {
t.Fatal("expected error for missing file, got nil")
}
}

func TestLoad_Valid(t *testing.T) {
os.Setenv("TEST_MIRDAIN_PAT", "ghp_abc123")
t.Cleanup(func() { os.Unsetenv("TEST_MIRDAIN_PAT") })

content := []byte(`
server:
  bind: 127.0.0.1
  port: 7777
github_app:
  pat: "${TEST_MIRDAIN_PAT}"
repos:
  - id: github.com/owner/repo
    enabled: true
`)
f, _ := os.CreateTemp("", "mirdain*.yaml")
f.Write(content)
f.Close()
defer os.Remove(f.Name())

cfg, err := config.Load(f.Name())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.Server.Bind != "127.0.0.1" {
t.Errorf("bind = %q, want 127.0.0.1", cfg.Server.Bind)
}
if cfg.Server.Port != 7777 {
t.Errorf("port = %d, want 7777", cfg.Server.Port)
}
if cfg.GitHubApp.PAT != "ghp_abc123" {
t.Errorf("PAT = %q, want ghp_abc123", cfg.GitHubApp.PAT)
}
if len(cfg.Repos) != 1 || cfg.Repos[0].ID != "github.com/owner/repo" {
t.Errorf("repos[0].id unexpected: %+v", cfg.Repos)
}
}

func TestLoad_Defaults(t *testing.T) {
content := []byte("repos:\n  - id: github.com/a/b\n    enabled: true\n")
f, _ := os.CreateTemp("", "mirdain*.yaml")
f.Write(content)
f.Close()
defer os.Remove(f.Name())

cfg, err := config.Load(f.Name())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.Server.Bind != "127.0.0.1" {
t.Errorf("default bind = %q, want 127.0.0.1", cfg.Server.Bind)
}
if cfg.Server.Port != 7777 {
t.Errorf("default port = %d, want 7777", cfg.Server.Port)
}
}
