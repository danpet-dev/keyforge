package commands

import (
	"os"
	"strings"
	"testing"
)

func TestRemoveMemberCommand(t *testing.T) {
	t.Run("no flags provided", func(t *testing.T) {
		// Save and restore flags
		origEmail := removeMemberEmail
		origFingerprint := removeMemberFingerprint
		origPublicKey := removeMemberPublicKey
		defer func() {
			removeMemberEmail = origEmail
			removeMemberFingerprint = origFingerprint
			removeMemberPublicKey = origPublicKey
		}()

		removeMemberEmail = ""
		removeMemberFingerprint = ""
		removeMemberPublicKey = ""

		// Create temp .sops.yaml
		tempDir := t.TempDir()
		sopsFile := tempDir + "/.sops.yaml"
		content := `creation_rules:
  - path_regex: test\.yaml\.sops$
    pgp: KEY1
`
		if err := os.WriteFile(sopsFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp .sops.yaml: %v", err)
		}

		// Change to temp dir
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tempDir)

		err := runRemoveMember(nil, []string{})

		if err == nil {
			t.Error("Expected error when no flags provided")
		}

		if !strings.Contains(err.Error(), "must provide one of") {
			t.Errorf("Expected 'must provide one of' error, got: %v", err)
		}
	})

	t.Run("multiple flags provided", func(t *testing.T) {
		origEmail := removeMemberEmail
		origFingerprint := removeMemberFingerprint
		defer func() {
			removeMemberEmail = origEmail
			removeMemberFingerprint = origFingerprint
		}()

		removeMemberEmail = "test@example.com"
		removeMemberFingerprint = "KEY123"
		removeMemberPublicKey = ""

		tempDir := t.TempDir()
		sopsFile := tempDir + "/.sops.yaml"
		content := `creation_rules:
  - path_regex: test\.yaml\.sops$
    pgp: KEY1
`
		if err := os.WriteFile(sopsFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp .sops.yaml: %v", err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tempDir)

		err := runRemoveMember(nil, []string{})

		if err == nil {
			t.Error("Expected error when multiple flags provided")
		}

		if !strings.Contains(err.Error(), "provide only one of") {
			t.Errorf("Expected 'provide only one of' error, got: %v", err)
		}
	})

	t.Run(".sops.yaml not found", func(t *testing.T) {
		origFingerprint := removeMemberFingerprint
		defer func() { removeMemberFingerprint = origFingerprint }()

		removeMemberEmail = ""
		removeMemberFingerprint = "KEY123"
		removeMemberPublicKey = ""

		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tempDir)

		err := runRemoveMember(nil, []string{})

		if err == nil {
			t.Error("Expected error when .sops.yaml not found")
		}

		if !strings.Contains(err.Error(), ".sops.yaml not found") {
			t.Errorf("Expected '.sops.yaml not found' error, got: %v", err)
		}
	})

	t.Run("key not found in .sops.yaml", func(t *testing.T) {
		origFingerprint := removeMemberFingerprint
		defer func() { removeMemberFingerprint = origFingerprint }()

		removeMemberEmail = ""
		removeMemberFingerprint = "NONEXISTENT_KEY"
		removeMemberPublicKey = ""

		tempDir := t.TempDir()
		sopsFile := tempDir + "/.sops.yaml"
		content := `creation_rules:
  - path_regex: test\.yaml\.sops$
    pgp: KEY1
`
		if err := os.WriteFile(sopsFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp .sops.yaml: %v", err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tempDir)

		err := runRemoveMember(nil, []string{})

		if err == nil {
			t.Error("Expected error when key not found")
		}

		if !strings.Contains(err.Error(), "not found in .sops.yaml") {
			t.Errorf("Expected 'not found in .sops.yaml' error, got: %v", err)
		}
	})

	t.Run("help shows correct usage", func(t *testing.T) {
		if removeMemberCmd.Use != "remove-member" {
			t.Errorf("Expected use='remove-member', got %s", removeMemberCmd.Use)
		}

		if !strings.Contains(removeMemberCmd.Short, "Remove") {
			t.Error("Short description missing 'Remove'")
		}
	})
}
