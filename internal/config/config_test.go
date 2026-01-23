package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write test config
	content := []byte("url: https://test.example.com\ntoken: test-token-123\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify values
	if cfg.URL != "https://test.example.com" {
		t.Errorf("expected URL 'https://test.example.com', got '%s'", cfg.URL)
	}
	if cfg.Token != "test-token-123" {
		t.Errorf("expected Token 'test-token-123', got '%s'", cfg.Token)
	}
}

func TestLoad_EnvVarOverride(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write test config
	content := []byte("url: https://file.example.com\ntoken: file-token\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Set environment variables
	os.Setenv("LINKDING_URL", "https://env.example.com")
	os.Setenv("LINKDING_TOKEN", "env-token")
	defer func() {
		os.Unsetenv("LINKDING_URL")
		os.Unsetenv("LINKDING_TOKEN")
	}()

	// Load config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify env vars took precedence
	if cfg.URL != "https://env.example.com" {
		t.Errorf("expected URL from env 'https://env.example.com', got '%s'", cfg.URL)
	}
	if cfg.Token != "env-token" {
		t.Errorf("expected Token from env 'env-token', got '%s'", cfg.Token)
	}
}

func TestLoad_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}

	expectedMsg := "no configuration found"
	if err.Error() != expectedMsg+". Run 'ld config init' to set up" {
		t.Errorf("expected error containing '%s', got '%v'", expectedMsg, err)
	}
}

func TestLoad_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "missing token",
			content: "url: https://test.example.com\n",
		},
		{
			name:    "missing url",
			content: "token: test-token\n",
		},
		{
			name:    "empty file",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			_, err := Load(configPath)
			if err == nil {
				t.Fatal("expected error for missing required fields, got nil")
			}
		})
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		URL:   "https://save.example.com",
		Token: "save-token-456",
	}

	// Save config
	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load config back and verify
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loaded.URL != cfg.URL {
		t.Errorf("expected URL '%s', got '%s'", cfg.URL, loaded.URL)
	}
	if loaded.Token != cfg.Token {
		t.Errorf("expected Token '%s', got '%s'", cfg.Token, loaded.Token)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "dir", "config.yaml")

	cfg := &Config{
		URL:   "https://test.example.com",
		Token: "test-token",
	}

	// Save config (should create nested directories)
	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created in nested directory")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath() failed: %v", err)
	}

	if path == "" {
		t.Error("expected non-empty path")
	}

	// Verify path contains expected components
	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}

	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected path to end with 'config.yaml', got '%s'", path)
	}
}
