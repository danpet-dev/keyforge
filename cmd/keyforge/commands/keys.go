package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/danpet-dev/keyforge/pkg/keys"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage encryption keys",
	Long: `Manage PGP and Age encryption keys.

This command provides subcommands for:
  - Listing available keys
  - Viewing key details
  - Exporting keys

Example:
  keyforge keys list
  keyforge keys list --type pgp
  keyforge keys list --format json`,
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available encryption keys",
	Long: `List all available PGP and Age encryption keys.

This command shows:
  - All PGP keys from GPG keyring
  - All Age keys from ~/.config/sops/age/keys.txt
  - Key fingerprints, names, emails, expiration dates
  - Warnings for expired or expiring keys

Example:
  keyforge keys list
  keyforge keys list --type pgp
  keyforge keys list --type age
  keyforge keys list --format json`,
	RunE: runKeysList,
}

var (
	keysListType   string
	keysListFormat string
)

func init() {
	rootCmd.AddCommand(keysCmd)
	keysCmd.AddCommand(keysListCmd)

	keysListCmd.Flags().StringVarP(&keysListType, "type", "t", "", "Filter by key type: pgp or age")
	keysListCmd.Flags().StringVarP(&keysListFormat, "format", "f", "text", "Output format: text or json")
}

func runKeysList(cmd *cobra.Command, args []string) error {
	var allKeys []keys.Key

	// Load PGP keys
	if keysListType == "" || keysListType == "pgp" {
		if keys.IsGPGInstalled() {
			pgpKeys, err := keys.ListPGPKeys()
			if err != nil {
				return fmt.Errorf("failed to list PGP keys: %w", err)
			}
			allKeys = append(allKeys, pgpKeys...)
		} else if keysListType == "pgp" {
			return fmt.Errorf("GPG is not installed. Please install GPG to manage PGP keys")
		}
	}

	// Load Age keys
	if keysListType == "" || keysListType == "age" {
		ageKeys, err := keys.ListAgeKeys()
		if err != nil {
			// Don't fail if age keys file doesn't exist
			if keysListType == "age" {
				return fmt.Errorf("failed to list Age keys: %w", err)
			}
		} else {
			allKeys = append(allKeys, ageKeys...)
		}
	}

	// Validate type filter
	if keysListType != "" && keysListType != "pgp" && keysListType != "age" {
		return fmt.Errorf("invalid key type: %s (must be 'pgp' or 'age')", keysListType)
	}

	// Check if any keys found
	if len(allKeys) == 0 {
		if keysListType == "pgp" {
			return fmt.Errorf("no PGP keys found in GPG keyring. Run 'keyforge init --generate-key' to create one")
		} else if keysListType == "age" {
			return fmt.Errorf("no Age keys found in ~/.config/sops/age/keys.txt")
		}
		return fmt.Errorf("no encryption keys found")
	}

	// Output
	if keysListFormat == "json" {
		return outputKeysJSON(allKeys)
	}

	return outputKeysText(allKeys)
}

func outputKeysJSON(allKeys []keys.Key) error {
	type JSONKey struct {
		Type        string  `json:"type"`
		Fingerprint string  `json:"fingerprint,omitempty"`
		PublicKey   string  `json:"public_key,omitempty"`
		Name        string  `json:"name,omitempty"`
		Email       string  `json:"email,omitempty"`
		Expires     *string `json:"expires,omitempty"`
		Warning     string  `json:"warning,omitempty"`
	}

	jsonKeys := make([]JSONKey, 0, len(allKeys))
	for _, key := range allKeys {
		jk := JSONKey{
			Type:  key.Type,
			Name:  key.Name,
			Email: key.Email,
		}

		if key.Type == "pgp" {
			jk.Fingerprint = key.Fingerprint
			if key.Expires != nil {
				expiresStr := key.Expires.Format("2006-01-02")
				jk.Expires = &expiresStr

				// Check expiration
				isExpiring, msg := keys.CheckKeyExpiration(key)
				if isExpiring {
					jk.Warning = msg
				}
			}
		} else {
			jk.PublicKey = key.PublicKey
		}

		jsonKeys = append(jsonKeys, jk)
	}

	output := map[string]interface{}{
		"total": len(jsonKeys),
		"keys":  jsonKeys,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func outputKeysText(allKeys []keys.Key) error {
	fmt.Println("🔑 Available Encryption Keys")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("\nTotal keys: %d\n\n", len(allKeys))

	// Group by type
	pgpKeys := []keys.Key{}
	ageKeys := []keys.Key{}

	for _, key := range allKeys {
		if key.Type == "pgp" {
			pgpKeys = append(pgpKeys, key)
		} else {
			ageKeys = append(ageKeys, key)
		}
	}

	// Display PGP keys
	if len(pgpKeys) > 0 {
		fmt.Printf("PGP Keys (%d):\n", len(pgpKeys))
		fmt.Println(strings.Repeat("-", 70))

		for _, key := range pgpKeys {
			fmt.Printf("\n📌 Fingerprint: %s\n", key.Fingerprint)
			if key.Name != "" {
				fmt.Printf("   Name:        %s", key.Name)
				if key.Email != "" {
					fmt.Printf(" <%s>", key.Email)
				}
				fmt.Println()
			}

			if key.Expires != nil {
				fmt.Printf("   Expires:     %s", key.Expires.Format("2006-01-02"))

				// Check expiration
				isExpiring, msg := keys.CheckKeyExpiration(key)
				if isExpiring {
					fmt.Printf(" ⚠️  %s", msg)
				}
				fmt.Println()
			} else {
				fmt.Println("   Expires:     never")
			}
		}
		fmt.Println()
	}

	// Display Age keys
	if len(ageKeys) > 0 {
		if len(pgpKeys) > 0 {
			fmt.Println()
		}

		fmt.Printf("Age Keys (%d):\n", len(ageKeys))
		fmt.Println(strings.Repeat("-", 70))

		for _, key := range ageKeys {
			fmt.Printf("\n📌 Public Key:  %s\n", key.PublicKey)
			if key.Name != "" {
				fmt.Printf("   Comment:     %s\n", key.Name)
			}
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 70))

	// Summary warnings
	hasWarnings := false
	for _, key := range pgpKeys {
		if key.Expires != nil {
			isExpiring, msg := keys.CheckKeyExpiration(key)
			if isExpiring {
				if !hasWarnings {
					fmt.Println("\n⚠️  Warnings:")
					hasWarnings = true
				}
				fmt.Printf("  - PGP key %s: %s\n", key.Fingerprint[:16]+"...", msg)
			}
		}
	}

	if hasWarnings {
		fmt.Println("\n💡 Tip: Consider rotating expiring keys with 'keyforge rotate-keys'")
	}

	return nil
}
