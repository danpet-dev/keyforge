package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStatusCommand(t *testing.T) {
	t.Run("no .sops.yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		chdirForTest(t, tempDir)

		// Should not error, just show warning
		err := runStatus(nil, []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("valid .sops.yaml with no files", func(t *testing.T) {
		tempDir := t.TempDir()
		chdirForTest(t, tempDir)

		content := `creation_rules:
  - path_regex: secrets/.*\.yaml\.sops$
    pgp: KEY1
`
		if err := os.WriteFile(".sops.yaml", []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to create .sops.yaml: %v", err)
		}

		err := runStatus(nil, []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("encrypted files found", func(t *testing.T) {
		tempDir := t.TempDir()
		chdirForTest(t, tempDir)

		content := `creation_rules:
  - path_regex: secrets/.*\.yaml\.sops$
    pgp: KEY1
`
		if err := os.WriteFile(".sops.yaml", []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to create .sops.yaml: %v", err)
		}

		// Create encrypted files
		if err := os.MkdirAll("secrets", 0o755); err != nil {
			t.Fatalf("Failed to create secrets directory: %v", err)
		}
		if err := os.WriteFile("secrets/dev.yaml.sops", []byte("encrypted"), 0o644); err != nil {
			t.Fatalf("Failed to write encrypted file: %v", err)
		}

		err := runStatus(nil, []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("help shows correct usage", func(t *testing.T) {
		if statusCmd.Use != "status" {
			t.Errorf("Expected use='status', got %s", statusCmd.Use)
		}

		if statusCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})
}

func TestFindDecryptedFiles(t *testing.T) {
	t.Run("no encrypted files", func(t *testing.T) {
		result := findDecryptedFiles([]string{})
		if len(result) != 0 {
			t.Errorf("Expected empty result, got %v", result)
		}
	})

	t.Run("with decrypted file", func(t *testing.T) {
		tempDir := t.TempDir()

		encFile := filepath.Join(tempDir, "test.yaml.sops")
		plainFile := filepath.Join(tempDir, "test.yaml")

		// Create both files
		if err := os.WriteFile(encFile, []byte("encrypted"), 0o644); err != nil {
			t.Fatalf("Failed to write encrypted test file: %v", err)
		}
		if err := os.WriteFile(plainFile, []byte("plain"), 0o644); err != nil {
			t.Fatalf("Failed to write plaintext test file: %v", err)
		}

		result := findDecryptedFiles([]string{encFile})
		if len(result) != 1 {
			t.Errorf("Expected 1 decrypted file, got %d", len(result))
		}
		if result[0] != plainFile {
			t.Errorf("Expected %s, got %s", plainFile, result[0])
		}
	})

	t.Run("without decrypted file", func(t *testing.T) {
		tempDir := t.TempDir()

		encFile := filepath.Join(tempDir, "test.yaml.sops")
		if err := os.WriteFile(encFile, []byte("encrypted"), 0o644); err != nil {
			t.Fatalf("Failed to write encrypted test file: %v", err)
		}

		result := findDecryptedFiles([]string{encFile})
		if len(result) != 0 {
			t.Errorf("Expected no decrypted files, got %v", result)
		}
	})
}

func TestCheckGitignore(t *testing.T) {
	t.Run("no decrypted files", func(t *testing.T) {
		result := checkGitignore([]string{})
		if len(result) != 0 {
			t.Errorf("Expected empty result, got %v", result)
		}
	})

	t.Run("no .gitignore file", func(t *testing.T) {
		tempDir := t.TempDir()
		chdirForTest(t, tempDir)

		decrypted := []string{"secrets/dev.yaml"}
		result := checkGitignore(decrypted)

		if len(result) != 1 {
			t.Errorf("Expected 1 issue, got %d", len(result))
		}
	})

	t.Run("file in .gitignore", func(t *testing.T) {
		tempDir := t.TempDir()
		chdirForTest(t, tempDir)

		// Create .gitignore
		gitignoreContent := "secrets/dev.yaml\n*.log\n"
		if err := os.WriteFile(".gitignore", []byte(gitignoreContent), 0o644); err != nil {
			t.Fatalf("Failed to write .gitignore: %v", err)
		}

		decrypted := []string{"secrets/dev.yaml"}
		result := checkGitignore(decrypted)

		if len(result) != 0 {
			t.Errorf("Expected no issues, got %v", result)
		}
	})

	t.Run("file not in .gitignore", func(t *testing.T) {
		tempDir := t.TempDir()
		chdirForTest(t, tempDir)

		// Create .gitignore
		gitignoreContent := "*.log\n"
		if err := os.WriteFile(".gitignore", []byte(gitignoreContent), 0o644); err != nil {
			t.Fatalf("Failed to write .gitignore: %v", err)
		}

		decrypted := []string{"secrets/dev.yaml"}
		result := checkGitignore(decrypted)

		if len(result) != 1 {
			t.Errorf("Expected 1 issue, got %d", len(result))
		}
	})
}

func TestIsIgnored(t *testing.T) {
	patterns := []string{"*.log", "secrets/dev.yaml", "temp"}

	tests := []struct {
		file     string
		expected bool
	}{
		{"test.log", true},
		{"secrets/dev.yaml", true},
		{"temp/file.txt", true},
		{"secrets/prod.yaml", false},
		{"other.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			result := isIgnored(tt.file, patterns)
			if result != tt.expected {
				t.Errorf("isIgnored(%s) = %v, want %v", tt.file, result, tt.expected)
			}
		})
	}
}
