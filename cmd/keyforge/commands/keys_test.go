package commands

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/danpet-dev/keyforge/pkg/keys"
)

func TestKeysListCommand(t *testing.T) {
	t.Run("invalid type filter", func(t *testing.T) {
		// Save and restore
		original := keysListType
		defer func() { keysListType = original }()

		keysListType = "invalid"

		err := runKeysList(nil, []string{})

		if err == nil {
			t.Error("Expected error for invalid type filter")
		}

		if !strings.Contains(err.Error(), "invalid key type") {
			t.Errorf("Expected 'invalid key type' error, got: %v", err)
		}
	})

	t.Run("help shows correct usage", func(t *testing.T) {
		if keysListCmd.Use != "list" {
			t.Errorf("Expected use='list', got %s", keysListCmd.Use)
		}

		if !strings.Contains(keysListCmd.Short, "encryption keys") {
			t.Error("Short description missing 'encryption keys'")
		}
	})
}

func TestOutputKeysJSON(t *testing.T) {
	mockKeys := []keys.Key{
		{
			Type:        "pgp",
			Fingerprint: "TEST1234",
			Name:        "Test",
			Email:       "test@example.com",
		},
		{
			Type:      "age",
			PublicKey: "age123",
			Name:      "Age Test",
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputKeysJSON(mockKeys)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputKeysJSON failed: %v", err)
	}

	output := make([]byte, 1024)
	n, _ := r.Read(output)
	outputStr := string(output[:n])

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(outputStr), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if total, ok := result["total"].(float64); !ok || total != 2 {
		t.Errorf("Expected total=2, got %v", result["total"])
	}

	keysArray, ok := result["keys"].([]interface{})
	if !ok || len(keysArray) != 2 {
		t.Errorf("Expected 2 keys in JSON array, got %v", result["keys"])
	}
}

func TestOutputKeysText(t *testing.T) {
	mockKeys := []keys.Key{
		{
			Type:        "pgp",
			Fingerprint: "TEST1234",
			Name:        "Test",
			Email:       "test@example.com",
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputKeysText(mockKeys)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputKeysText failed: %v", err)
	}

	output := make([]byte, 1024)
	n, _ := r.Read(output)
	outputStr := string(output[:n])

	if !strings.Contains(outputStr, "Available Encryption Keys") {
		t.Error("Output missing header")
	}

	if !strings.Contains(outputStr, "TEST1234") {
		t.Error("Output missing fingerprint")
	}

	if !strings.Contains(outputStr, "Test") {
		t.Error("Output missing name")
	}
}
