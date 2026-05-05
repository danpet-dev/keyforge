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

	pgpKeys := cfg.CreationRules[0].GetPGPKeys()
	if len(pgpKeys) != 2 {
		t.Errorf("Expected 2 PGP keys in development rule, got %d", len(pgpKeys))
	}
}

func TestLoadSave(t *testing.T) {
	tempFile := "/tmp/test-sops.yaml"
	defer func() {
		if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to remove temp file %s: %v", tempFile, err)
		}
	}()

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

	pgpKeys := cfg.CreationRules[0].GetPGPKeys()
	if len(pgpKeys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(pgpKeys))
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

func TestRemoveKey(t *testing.T) {
	t.Run("remove PGP key from single rule", func(t *testing.T) {
		cfg := &SopsConfig{
			CreationRules: []CreationRule{
				{
					PathRegex: "secrets/dev\\.yaml\\.sops$",
					PGP:       []string{"KEY1", "KEY2"},
				},
			},
		}

		affected := cfg.RemoveKey("pgp", "KEY1")

		if len(affected) != 1 {
			t.Errorf("Expected 1 affected rule, got %d", len(affected))
		}

		pgpKeys := cfg.CreationRules[0].GetPGPKeys()
		if len(pgpKeys) != 1 || pgpKeys[0] != "KEY2" {
			t.Errorf("Expected KEY2 only, got %v", pgpKeys)
		}
	})

	t.Run("remove PGP key from multiple rules", func(t *testing.T) {
		cfg := &SopsConfig{
			CreationRules: []CreationRule{
				{
					PathRegex: "secrets/dev\\.yaml\\.sops$",
					PGP:       []string{"KEY1", "KEY2"},
				},
				{
					PathRegex: "secrets/prod\\.yaml\\.sops$",
					PGP:       []string{"KEY1"},
				},
			},
		}

		affected := cfg.RemoveKey("pgp", "KEY1")

		if len(affected) != 2 {
			t.Errorf("Expected 2 affected rules, got %d", len(affected))
		}

		// Check first rule
		pgpKeys1 := cfg.CreationRules[0].GetPGPKeys()
		if len(pgpKeys1) != 1 || pgpKeys1[0] != "KEY2" {
			t.Errorf("Expected KEY2 in first rule, got %v", pgpKeys1)
		}

		// Check second rule - should have no keys
		pgpKeys2 := cfg.CreationRules[1].GetPGPKeys()
		if len(pgpKeys2) != 0 {
			t.Errorf("Expected no keys in second rule, got %v", pgpKeys2)
		}
	})

	t.Run("remove Age key", func(t *testing.T) {
		cfg := &SopsConfig{
			CreationRules: []CreationRule{
				{
					PathRegex: "secrets/dev\\.yaml\\.sops$",
					Age:       []string{"age123", "age456"},
				},
			},
		}

		affected := cfg.RemoveKey("age", "age123")

		if len(affected) != 1 {
			t.Errorf("Expected 1 affected rule, got %d", len(affected))
		}

		ageKeys := cfg.CreationRules[0].GetAgeKeys()
		if len(ageKeys) != 1 || ageKeys[0] != "age456" {
			t.Errorf("Expected age456 only, got %v", ageKeys)
		}
	})

	t.Run("remove non-existent key", func(t *testing.T) {
		cfg := &SopsConfig{
			CreationRules: []CreationRule{
				{
					PathRegex: "secrets/dev\\.yaml\\.sops$",
					PGP:       []string{"KEY1"},
				},
			},
		}

		affected := cfg.RemoveKey("pgp", "KEY_NONEXISTENT")

		if len(affected) != 0 {
			t.Errorf("Expected 0 affected rules, got %d", len(affected))
		}
	})
}

func TestNormalizeKeys(t *testing.T) {
	t.Run("comma-separated string", func(t *testing.T) {
		input := "KEY1, KEY2, KEY3"
		result := normalizeKeys(input)

		if len(result) != 3 {
			t.Errorf("Expected 3 keys, got %d", len(result))
		}

		expected := []string{"KEY1", "KEY2", "KEY3"}
		for i, key := range expected {
			if result[i] != key {
				t.Errorf("Expected %s at index %d, got %s", key, i, result[i])
			}
		}
	})

	t.Run("comma-separated with newlines", func(t *testing.T) {
		input := "KEY1,\nKEY2,\nKEY3"
		result := normalizeKeys(input)

		if len(result) != 3 {
			t.Errorf("Expected 3 keys, got %d", len(result))
		}
	})

	t.Run("single string without comma", func(t *testing.T) {
		input := "SINGLE_KEY"
		result := normalizeKeys(input)

		if len(result) != 1 || result[0] != "SINGLE_KEY" {
			t.Errorf("Expected [SINGLE_KEY], got %v", result)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		input := []string{"KEY1", "KEY2"}
		result := normalizeKeys(input)

		if len(result) != 2 {
			t.Errorf("Expected 2 keys, got %d", len(result))
		}
	})

	t.Run("interface slice", func(t *testing.T) {
		input := []interface{}{"KEY1", "KEY2"}
		result := normalizeKeys(input)

		if len(result) != 2 {
			t.Errorf("Expected 2 keys, got %d", len(result))
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result := normalizeKeys(nil)

		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})
}
