package keys

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Key represents a PGP or Age key
type Key struct {
	Type        string    // "pgp" or "age"
	Fingerprint string    // For PGP
	PublicKey   string    // For Age
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
