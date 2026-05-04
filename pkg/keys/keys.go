package keys

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"
)

// Key represents a PGP or Age key
type Key struct {
	Type        string // "pgp" or "age"
	Fingerprint string // For PGP
	PublicKey   string // For Age
	Name        string
	Email       string
	Created     time.Time
	Expires     *time.Time
}

// ListPGPKeys lists all PGP keys in the keyring
func ListPGPKeys() ([]Key, error) {
	cmd := exec.Command("gpg", "--list-keys", "--with-colons")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list gpg keys: %w", err)
	}

	var keys []Key
	var currentKey *Key

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 10 {
			continue
		}

		switch fields[0] {
		case "pub":
			if currentKey != nil {
				keys = append(keys, *currentKey)
			}
			currentKey = &Key{
				Type:        "pgp",
				Fingerprint: fields[4],
			}
			// Parse creation date
			if created, err := parseTimestamp(fields[5]); err == nil {
				currentKey.Created = created
			}
			// Parse expiration date
			if fields[6] != "" {
				if expires, err := parseTimestamp(fields[6]); err == nil {
					currentKey.Expires = &expires
				}
			}

		case "uid":
			if currentKey != nil && len(fields) > 9 {
				// Parse "Name <email>"
				uid := fields[9]
				if strings.Contains(uid, "<") && strings.Contains(uid, ">") {
					parts := strings.Split(uid, "<")
					currentKey.Name = strings.TrimSpace(parts[0])
					currentKey.Email = strings.TrimSuffix(strings.TrimPrefix(parts[1], "<"), ">")
				} else {
					currentKey.Name = uid
				}
			}
		}
	}

	if currentKey != nil {
		keys = append(keys, *currentKey)
	}

	return keys, nil
}

// GeneratePGPKey generates a new PGP key
func GeneratePGPKey(name, email string, years int) (string, error) {
	// Create GPG key generation batch file
	batch := fmt.Sprintf(`%%echo Generating PGP key
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: %s
Name-Email: %s
Expire-Date: %dy
%%no-protection
%%commit
%%echo done`, name, email, years)

	cmd := exec.Command("gpg", "--batch", "--generate-key")
	cmd.Stdin = strings.NewReader(batch)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to generate PGP key: %s: %w", stderr.String(), err)
	}

	// Get the fingerprint of the newly created key
	keys, err := ListPGPKeys()
	if err != nil {
		return "", fmt.Errorf("failed to list keys after generation: %w", err)
	}

	for _, key := range keys {
		if key.Email == email {
			return key.Fingerprint, nil
		}
	}

	return "", fmt.Errorf("failed to find newly generated key")
}

// ExportPublicKey exports the public key for a given fingerprint
func ExportPublicKey(fingerprint string) (string, error) {
	cmd := exec.Command("gpg", "--armor", "--export", fingerprint)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to export public key: %w", err)
	}

	return string(output), nil
}

// CheckKeyExpiration checks if a key is expired or expiring soon (within 30 days)
func CheckKeyExpiration(key Key) (bool, string) {
	if key.Expires == nil {
		return false, "Key does not expire"
	}

	now := time.Now()
	daysUntilExpiry := int(key.Expires.Sub(now).Hours() / 24)

	if now.After(*key.Expires) {
		return true, fmt.Sprintf("Key expired %d days ago", -daysUntilExpiry)
	}

	if daysUntilExpiry <= 30 {
		return true, fmt.Sprintf("Key expires in %d days", daysUntilExpiry)
	}

	return false, fmt.Sprintf("Key expires in %d days", daysUntilExpiry)
}

// parseTimestamp parses a Unix timestamp string
func parseTimestamp(ts string) (time.Time, error) {
	if ts == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}

	var sec int64
	_, err := fmt.Sscanf(ts, "%d", &sec)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(sec, 0), nil
}

// IsGPGInstalled checks if GPG is available in PATH
func IsGPGInstalled() bool {
	_, err := exec.LookPath("gpg")
	return err == nil
}

// --- Age Key Functions ---

// GenerateAgeKey generates a new Age key pair
func GenerateAgeKey() (*age.X25519Identity, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generate age key: %w", err)
	}

	return identity, nil
}

// SaveAgeKey saves an Age key to the default location (~/.config/sops/age/keys.txt)
func SaveAgeKey(identity *age.X25519Identity, comment string) (err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	ageDir := filepath.Join(homeDir, ".config", "sops", "age")
	if err := os.MkdirAll(ageDir, 0700); err != nil {
		return fmt.Errorf("failed to create age directory: %w", err)
	}

	keysFile := filepath.Join(ageDir, "keys.txt")

	// Check if file exists
	var existingContent string
	if data, err := os.ReadFile(keysFile); err == nil {
		existingContent = string(data)
	}

	// Append new key
	f, err := os.OpenFile(keysFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open keys file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close keys file: %w", cerr)
		}
	}()

	// Add comment if provided
	if comment != "" {
		if existingContent != "" && !strings.HasSuffix(existingContent, "\n") {
			if _, err := f.WriteString("\n"); err != nil {
				return fmt.Errorf("failed writing newline to keys file: %w", err)
			}
		}
		if _, err := f.WriteString(fmt.Sprintf("# %s\n", comment)); err != nil {
			return fmt.Errorf("failed writing comment to keys file: %w", err)
		}
	}

	// Write private key
	if _, err := f.WriteString(identity.String() + "\n"); err != nil {
		return fmt.Errorf("failed writing age identity to keys file: %w", err)
	}

	return nil
}

// ListAgeKeys lists Age keys from ~/.config/sops/age/keys.txt
func ListAgeKeys() ([]Key, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	keysFile := filepath.Join(homeDir, ".config", "sops", "age", "keys.txt")
	data, err := os.ReadFile(keysFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Key{}, nil
		}
		return nil, fmt.Errorf("failed to read age keys: %w", err)
	}

	var keys []Key
	var currentComment string

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			currentComment = strings.TrimPrefix(line, "# ")
			continue
		}

		// Parse Age private key
		if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
			identity, err := age.ParseX25519Identity(line)
			if err != nil {
				continue
			}

			key := Key{
				Type:      "age",
				PublicKey: identity.Recipient().String(),
				Name:      currentComment,
			}
			keys = append(keys, key)
			currentComment = ""
		}
	}

	return keys, nil
}

// GetAgePublicKey returns the public key for an Age private key
func GetAgePublicKey(privateKey string) (string, error) {
	identity, err := age.ParseX25519Identity(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse age key: %w", err)
	}

	return identity.Recipient().String(), nil
}

// IsAgeInstalled checks if age is available in PATH
func IsAgeInstalled() bool {
	_, err := exec.LookPath("age")
	return err == nil
}
