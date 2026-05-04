package config

import (
	"os"
	"testing"
)

func TestGenerateTemplate(t *testing.T) {
	keys := map[string][]string{
		"development": {"KEY1", "KEY2"},
		"test":        {"KEY1"},
		"prod":        {"KEY1"},
		"all":         {"KEY1"},
	}

	cfg := GenerateTemplate("test-project", "pgp", keys)

	if len(cfg.CreationRules) != 5 {
		t.Errorf("Expected 5 creation rules, got %d", len(cfg.CreationRules))
	}

	// Check first rule
	if cfg.CreationRules[0].PathRegex != "secrets/development\\.yaml\\.sops$" {
		t.Errorf("Unexpected path regex: %s", cfg.CreationRules[0].PathRegex)
	}

	if len(cfg.CreationRules[0].PGP) != 2 {
		t.Errorf("Expected 2 PGP keys in development rule, got %d", len(cfg.CreationRules[0].PGP))
	}
}

func TestLoadSave(t *testing.T) {
	tempFile := "/tmp/test-sops.yaml"
	defer os.Remove(tempFile)

	// Create test config
	cfg := &SopsConfig{
		CreationRules: []CreationRule{
			{
				PathRegex:      "test\\.yaml\\.sops$",
				PGP:            []string{"TEST_KEY"},
				EncryptedRegex: "^(password|secret)$",
			},
		},
	}

	// Save
	if err := cfg.Save(tempFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load
	loadedCfg, err := Load(tempFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify
	if len(loadedCfg.CreationRules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(loadedCfg.CreationRules))
	}

	if loadedCfg.CreationRules[0].PathRegex != "test\\.yaml\\.sops$" {
		t.Errorf("Path regex mismatch: %s", loadedCfg.CreationRules[0].PathRegex)
	}
}

func TestAddKey(t *testing.T) {
	cfg := &SopsConfig{
		CreationRules: []CreationRule{
			{
				PathRegex: "secrets/.*\\.yaml\\.sops$",
				PGP:       []string{"KEY1"},
			},
		},
	}

	err := cfg.AddKey("secrets/.*\\.yaml\\.sops$", "pgp", "KEY2")
	if err != nil {
		t.Fatalf("Failed to add key: %v", err)
	}

	if len(cfg.CreationRules[0].PGP) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(cfg.CreationRules[0].PGP))
	}

	// Test non-existent path
	err = cfg.AddKey("nonexistent", "pgp", "KEY3")
	if err == nil {
		t.Error("Expected error for non-existent path regex")
	}

	// Test invalid key type
	err = cfg.AddKey("secrets/.*\\.yaml\\.sops$", "invalid", "KEY4")
	if err == nil {
		t.Error("Expected error for invalid key type")
	}
}
