package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/danpet-dev/keyforge/pkg/config"
	"github.com/danpet-dev/keyforge/pkg/keys"
	"github.com/spf13/cobra"
)

var removeMemberCmd = &cobra.Command{
	Use:   "remove-member",
	Short: "Remove a team member from .sops.yaml",
	Long: `Remove a team member's encryption key from .sops.yaml.

This command:
  1. Removes the key from all environments in .sops.yaml
  2. Shows which environments are affected
  3. Warns that existing encrypted files can still be decrypted
  4. Suggests running 'keyforge rotate-keys' to fully revoke access

IMPORTANT: Removing a key from .sops.yaml does NOT prevent the member
from decrypting existing files! They can still decrypt files that were
encrypted with their key. To fully revoke access, you must:
  1. Remove the key with this command
  2. Run 'keyforge rotate-keys' to re-encrypt all files with new keys
  3. OR manually run 'sops updatekeys' on all affected files

Example:
  keyforge remove-member --email alice@example.com
  keyforge remove-member --fingerprint ABCD1234EFGH5678
  keyforge remove-member --public-key age1234567890abcdef`,
	RunE: runRemoveMember,
}

var (
	removeMemberEmail      string
	removeMemberFingerprint string
	removeMemberPublicKey   string
)

func init() {
	rootCmd.AddCommand(removeMemberCmd)

	removeMemberCmd.Flags().StringVarP(&removeMemberEmail, "email", "e", "", "Remove PGP key by email")
	removeMemberCmd.Flags().StringVarP(&removeMemberFingerprint, "fingerprint", "f", "", "Remove PGP key by fingerprint")
	removeMemberCmd.Flags().StringVarP(&removeMemberPublicKey, "public-key", "p", "", "Remove Age key by public key")
}

func runRemoveMember(cmd *cobra.Command, args []string) error {
	// Check if .sops.yaml exists
	if _, err := os.Stat(".sops.yaml"); os.IsNotExist(err) {
		return fmt.Errorf(".sops.yaml not found. Run 'keyforge init' first")
	}

	// Validate that exactly one identifier is provided
	provided := 0
	if removeMemberEmail != "" {
		provided++
	}
	if removeMemberFingerprint != "" {
		provided++
	}
	if removeMemberPublicKey != "" {
		provided++
	}

	if provided == 0 {
		return fmt.Errorf("must provide one of: --email, --fingerprint, or --public-key")
	}
	if provided > 1 {
		return fmt.Errorf("provide only one of: --email, --fingerprint, or --public-key")
	}

	// Load .sops.yaml
	cfg, err := config.Load(".sops.yaml")
	if err != nil {
		return fmt.Errorf("failed to load .sops.yaml: %w", err)
	}

	var keyToRemove string
	var keyType string
	var memberInfo string

	// Determine which key to remove
	if removeMemberEmail != "" {
		// Find PGP key by email
		pgpKeys, err := keys.ListPGPKeys()
		if err != nil {
			return fmt.Errorf("failed to list PGP keys: %w", err)
		}

		found := false
		for _, key := range pgpKeys {
			if key.Email == removeMemberEmail {
				keyToRemove = key.Fingerprint
				keyType = "pgp"
				memberInfo = fmt.Sprintf("%s <%s>", key.Name, key.Email)
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("no PGP key found for email: %s", removeMemberEmail)
		}
	} else if removeMemberFingerprint != "" {
		keyToRemove = removeMemberFingerprint
		keyType = "pgp"

		// Try to get member info
		pgpKeys, err := keys.ListPGPKeys()
		if err == nil {
			for _, key := range pgpKeys {
				if key.Fingerprint == removeMemberFingerprint {
					memberInfo = fmt.Sprintf("%s <%s>", key.Name, key.Email)
					break
				}
			}
		}
		if memberInfo == "" {
			memberInfo = fmt.Sprintf("Fingerprint: %s", removeMemberFingerprint)
		}
	} else if removeMemberPublicKey != "" {
		keyToRemove = removeMemberPublicKey
		keyType = "age"

		// Try to get member info
		ageKeys, err := keys.ListAgeKeys()
		if err == nil {
			for _, key := range ageKeys {
				if key.PublicKey == removeMemberPublicKey {
					memberInfo = key.Name
					break
				}
			}
		}
		if memberInfo == "" {
			memberInfo = fmt.Sprintf("Public Key: %s", removeMemberPublicKey)
		}
	}

	// Check if key exists in .sops.yaml
	keyExists := false
	for _, rule := range cfg.CreationRules {
		if keyType == "pgp" {
			for _, k := range rule.GetPGPKeys() {
				if k == keyToRemove {
					keyExists = true
					break
				}
			}
		} else {
			for _, k := range rule.GetAgeKeys() {
				if k == keyToRemove {
					keyExists = true
					break
				}
			}
		}
		if keyExists {
			break
		}
	}

	if !keyExists {
		return fmt.Errorf("key not found in .sops.yaml: %s", keyToRemove)
	}

	// Show member info
	fmt.Println("🔑 Removing Team Member")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Member:   %s\n", memberInfo)
	fmt.Printf("Key Type: %s\n", strings.ToUpper(keyType))
	if keyType == "pgp" {
		fmt.Printf("Fingerprint: %s\n", keyToRemove)
	} else {
		fmt.Printf("Public Key: %s\n", keyToRemove)
	}
	fmt.Println()

	// Remove key and get affected rules
	affectedRules := cfg.RemoveKey(keyType, keyToRemove)

	if len(affectedRules) == 0 {
		return fmt.Errorf("key not found in any rules")
	}

	// Show affected environments
	fmt.Printf("📋 Affected Environments (%d):\n", len(affectedRules))
	fmt.Println(strings.Repeat("-", 70))
	for pathRegex := range affectedRules {
		fmt.Printf("  - %s\n", pathRegex)
	}
	fmt.Println()

	// Save updated .sops.yaml
	if err := cfg.Save(".sops.yaml"); err != nil {
		return fmt.Errorf("failed to save .sops.yaml: %w", err)
	}

	fmt.Println("✓ Successfully removed key from .sops.yaml")
	fmt.Println()

	// Security warning
	fmt.Println("⚠️  SECURITY WARNING")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("The member can STILL DECRYPT existing encrypted files!")
	fmt.Println()
	fmt.Println("Why? Because the files were encrypted with their key.")
	fmt.Println("Removing the key from .sops.yaml only prevents future encryption.")
	fmt.Println()
	fmt.Println("To fully revoke access, you must re-encrypt all files:")
	fmt.Println()
	fmt.Println("  Option 1 (Recommended): Rotate all keys")
	fmt.Println("    keyforge rotate-keys")
	fmt.Println()
	fmt.Println("  Option 2: Update keys on existing files")
	fmt.Println("    sops updatekeys secrets/development.yaml.sops")
	fmt.Println("    sops updatekeys secrets/test.yaml.sops")
	fmt.Println("    sops updatekeys secrets/prod.yaml.sops")
	fmt.Println()
	fmt.Println("  Option 3: Use update-all (for all encrypted files)")
	fmt.Println("    keyforge update-all")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))

	return nil
}
