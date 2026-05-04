package keys

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAgeKey(t *testing.T) {
	identity, err := GenerateAgeKey()
	if err != nil {
		t.Fatalf("GenerateAgeKey failed: %v", err)
	}

	if identity == nil {
		t.Fatal("Generated identity is nil")
	}

	// Check that public key is valid format
	pubKey := identity.Recipient().String()
	if !strings.HasPrefix(pubKey, "age1") {
		t.Errorf("Invalid Age public key format: %s", pubKey)
	}

	// Check that private key is valid format
	privKey := identity.String()
	if !strings.HasPrefix(privKey, "AGE-SECRET-KEY-") {
		t.Errorf("Invalid Age private key format: %s", privKey)
	}
}

func TestSaveAndListAgeKeys(t *testing.T) {
	// Create temporary age keys directory
	tmpDir := t.TempDir()
	ageDir := filepath.Join(tmpDir, ".config", "sops", "age")
	if err := os.MkdirAll(ageDir, 0700); err != nil {
		t.Fatalf("Failed to create age directory: %v", err)
	}

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Fatalf("failed to restore HOME: %v", err)
		}
	}()

	// Generate and save a key
	identity, err := GenerateAgeKey()
	if err != nil {
		t.Fatalf("GenerateAgeKey failed: %v", err)
	}

	comment := "Test key"
	if err := SaveAgeKey(identity, comment); err != nil {
		t.Fatalf("SaveAgeKey failed: %v", err)
	}

	// List keys
	keys, err := ListAgeKeys()
	if err != nil {
		t.Fatalf("ListAgeKeys failed: %v", err)
	}

	if len(keys) != 1 {
		t.Fatalf("Expected 1 key, got %d", len(keys))
	}

	key := keys[0]
	if key.Type != "age" {
		t.Errorf("Expected type 'age', got '%s'", key.Type)
	}

	if key.PublicKey != identity.Recipient().String() {
		t.Errorf("Public key mismatch: expected %s, got %s", identity.Recipient().String(), key.PublicKey)
	}

	if key.Name != comment {
		t.Errorf("Comment mismatch: expected %s, got %s", comment, key.Name)
	}
}

func TestListAgeKeysEmptyDir(t *testing.T) {
	// Create temporary age keys directory (empty)
	tmpDir := t.TempDir()
	ageDir := filepath.Join(tmpDir, ".config", "sops", "age")
	if err := os.MkdirAll(ageDir, 0700); err != nil {
		t.Fatalf("Failed to create age directory: %v", err)
	}

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Fatalf("failed to restore HOME: %v", err)
		}
	}()

	// List keys (should return empty list, not error)
	keys, err := ListAgeKeys()
	if err != nil {
		t.Fatalf("ListAgeKeys failed: %v", err)
	}

	if len(keys) != 0 {
		t.Fatalf("Expected 0 keys, got %d", len(keys))
	}
}

func TestGetAgePublicKey(t *testing.T) {
	identity, err := GenerateAgeKey()
	if err != nil {
		t.Fatalf("GenerateAgeKey failed: %v", err)
	}

	privateKey := identity.String()
	expectedPublic := identity.Recipient().String()

	publicKey, err := GetAgePublicKey(privateKey)
	if err != nil {
		t.Fatalf("GetAgePublicKey failed: %v", err)
	}

	if publicKey != expectedPublic {
		t.Errorf("Public key mismatch: expected %s, got %s", expectedPublic, publicKey)
	}
}

func TestGetAgePublicKeyInvalid(t *testing.T) {
	_, err := GetAgePublicKey("invalid-key")
	if err == nil {
		t.Error("Expected error for invalid Age key, got nil")
	}
}
