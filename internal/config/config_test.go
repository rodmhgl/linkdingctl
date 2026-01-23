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

func TestSave_PermissionsVerification(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := &Config{
		URL:   "https://test.example.com",
		Token: "test-token",
	}

	// Save config
	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify directory permissions (should be 0700)
	dirPath := filepath.Dir(configPath)
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}

	dirMode := dirInfo.Mode().Perm()
	expectedDirMode := os.FileMode(0700)
	if dirMode != expectedDirMode {
		t.Errorf("expected directory permissions %v, got %v", expectedDirMode, dirMode)
	}

	// Verify file permissions (should be 0600)
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}

	fileMode := fileInfo.Mode().Perm()
	expectedFileMode := os.FileMode(0600)
	if fileMode != expectedFileMode {
		t.Errorf("expected file permissions %v, got %v", expectedFileMode, fileMode)
	}
}

func TestLoad_NonYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write invalid YAML content (should trigger parse error)
	invalidContent := []byte("not: valid: yaml: content: [unclosed")
	if err := os.WriteFile(configPath, invalidContent, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load config
	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}

	// Verify it's a parse error (not just "not found")
	expectedErrSubstring := "failed to read config file"
	if err.Error()[:len(expectedErrSubstring)] != expectedErrSubstring {
		t.Errorf("expected error containing '%s', got '%v'", expectedErrSubstring, err)
	}
}

func TestLoad_EnvVarsOnly(t *testing.T) {
	tmpDir := t.TempDir()
	// Point to a non-existent config file
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// Set environment variables
	os.Setenv("LINKDING_URL", "https://envonly.example.com")
	os.Setenv("LINKDING_TOKEN", "envonly-token")
	defer func() {
		os.Unsetenv("LINKDING_URL")
		os.Unsetenv("LINKDING_TOKEN")
	}()

	// Load config (should succeed with env vars only)
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed with env vars only: %v", err)
	}

	// Verify values from env
	if cfg.URL != "https://envonly.example.com" {
		t.Errorf("expected URL from env 'https://envonly.example.com', got '%s'", cfg.URL)
	}
	if cfg.Token != "envonly-token" {
		t.Errorf("expected Token from env 'envonly-token', got '%s'", cfg.Token)
	}
}
