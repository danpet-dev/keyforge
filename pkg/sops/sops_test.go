package sops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindSopsFiles(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create test files
	files := []string{
		"secrets/dev.yaml.sops",
		"secrets/prod.yaml.sops",
		"services/app/k8s/secrets/test.yaml.sops",
		"regular.yaml",
		"config.json",
	}

	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) failed: %v", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) failed: %v", fullPath, err)
		}
	}

	// Find SOPS files
	found, err := FindSopsFiles(tempDir)
	if err != nil {
		t.Fatalf("FindSopsFiles failed: %v", err)
	}

	// Should find 3 .sops files
	if len(found) != 3 {
		t.Errorf("Expected 3 .sops files, found %d", len(found))
	}

	// Verify each file ends with .sops
	for _, f := range found {
		if filepath.Ext(f) != ".sops" {
			t.Errorf("File doesn't end with .sops: %s", f)
		}
	}
}

func TestIsSopsInstalled(t *testing.T) {
	// This test depends on system environment
	installed := IsSopsInstalled()

	// Log result for debugging (test passes regardless)
	if installed {
		t.Log("SOPS is installed")
	} else {
		t.Log("SOPS is not installed")
	}
}

func TestGetSopsVersion(t *testing.T) {
	if !IsSopsInstalled() {
		t.Skip("SOPS not installed, skipping version test")
	}

	version, err := GetSopsVersion()
	if err != nil {
		t.Fatalf("GetSopsVersion failed: %v", err)
	}

	if version == "" {
		t.Error("Expected non-empty version string")
	}

	t.Logf("SOPS version: %s", version)
}

func TestUpdatekeysWithoutSops(t *testing.T) {
	// Test error handling when file doesn't end with .sops
	err := Updatekeys("regular.yaml")
	if err == nil {
		t.Error("Expected error for file without .sops extension")
	}
}
