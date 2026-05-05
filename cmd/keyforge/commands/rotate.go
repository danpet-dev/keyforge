package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/danpet-dev/keyforge/pkg/config"
	"github.com/danpet-dev/keyforge/pkg/keys"
	"github.com/danpet-dev/keyforge/pkg/sops"
	"github.com/spf13/cobra"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate-keys",
	Short: "Rotate encryption keys",
	Long: `Rotate encryption keys for improved security.

This command:
  1. Generates new keys (PGP or Age)
  2. Adds new keys to .sops.yaml alongside old ones
  3. Re-encrypts all files with both old and new keys
  4. Removes old keys from .sops.yaml
  5. Final re-encryption with only new keys

Example:
  keyforge rotate-keys --key-type pgp --environments all
  keyforge rotate-keys --key-type age --environments prod
  keyforge rotate-keys --old-key B016BDB3E04C2D186389850787070D91C80C0400`,
	RunE: runRotate,
}

var (
	rotateKeyType      string
	rotateEnvironments string
	rotateOldKey       string
	rotateName         string
	rotateEmail        string
)

func init() {
	rootCmd.AddCommand(rotateCmd)

	rotateCmd.Flags().StringVarP(&rotateKeyType, "key-type", "t", "pgp", "Key type: pgp or age")
	rotateCmd.Flags().StringVarP(&rotateEnvironments, "environments", "E", "all", "Environments to rotate (comma-separated or 'all')")
	rotateCmd.Flags().StringVarP(&rotateOldKey, "old-key", "o", "", "Specific old key to replace (fingerprint or public key)")
	rotateCmd.Flags().StringVarP(&rotateName, "name", "n", "", "Name for new key (required for PGP)")
	rotateCmd.Flags().StringVarP(&rotateEmail, "email", "e", "", "Email for new key (required for PGP)")
}

func runRotate(cmd *cobra.Command, args []string) error {
	// Check if .sops.yaml exists
	if _, err := os.Stat(".sops.yaml"); os.IsNotExist(err) {
		return fmt.Errorf(".sops.yaml not found. Run 'keyforge init' first")
	}

	// Validate key type
	if rotateKeyType != "pgp" && rotateKeyType != "age" {
		return fmt.Errorf("invalid key-type: %s (must be 'pgp' or 'age')", rotateKeyType)
	}

	// Check prerequisites
	if rotateKeyType == "pgp" {
		if !keys.IsGPGInstalled() {
			return fmt.Errorf("GPG is not installed. Please install GPG first")
		}
		if rotateName == "" || rotateEmail == "" {
			return fmt.Errorf("--name and --email are required for PGP key rotation")
		}
	}

	// Load .sops.yaml
	cfg, err := config.Load(".sops.yaml")
	if err != nil {
		return fmt.Errorf("failed to load .sops.yaml: %w", err)
	}

	fmt.Println("🔄 Starting key rotation...")

	// Step 1: Generate new key
	var newKey string
	if rotateKeyType == "pgp" {
		fmt.Printf("Generating new PGP key for %s (%s)...\n", rotateName, rotateEmail)
		fingerprint, err := keys.GeneratePGPKey(rotateName, rotateEmail, 2)
		if err != nil {
			return fmt.Errorf("failed to generate PGP key: %w", err)
		}
		newKey = fingerprint
		fmt.Printf("✓ Generated new PGP key: %s\n", newKey)
	} else if rotateKeyType == "age" {
		fmt.Printf("Generating new Age key...\n")
		identity, err := keys.GenerateAgeKey()
		if err != nil {
			return fmt.Errorf("failed to generate Age key: %w", err)
		}

		comment := "Rotated key - created by keyforge"
		if err := keys.SaveAgeKey(identity, comment); err != nil {
			return fmt.Errorf("failed to save Age key: %w", err)
		}

		newKey = identity.Recipient().String()
		fmt.Printf("✓ Generated new Age key: %s\n", newKey)
	}

	// Step 2: Identify old keys to replace
	var oldKeys []string
	if rotateOldKey != "" {
		oldKeys = []string{rotateOldKey}
	} else {
		// Auto-detect old keys from first rule
		for _, rule := range cfg.CreationRules {
			if rotateKeyType == "pgp" {
				oldKeys = rule.GetPGPKeys()
			} else {
				oldKeys = rule.GetAgeKeys()
			}
			if len(oldKeys) > 0 {
				break
			}
		}
	}

	if len(oldKeys) == 0 {
		return fmt.Errorf("no old keys found to replace")
	}

	fmt.Printf("🔑 Replacing %d old key(s)\n", len(oldKeys))

	// Parse environments
	envs := strings.Split(rotateEnvironments, ",")
	if rotateEnvironments == "all" {
		envs = []string{"all"}
	}

	// Step 3: Add new key alongside old ones
	rulesUpdated := 0
	for i := range cfg.CreationRules {
		rule := &cfg.CreationRules[i]

		// Check if this rule matches requested environments
		shouldUpdate := false
		if rotateEnvironments == "all" {
			shouldUpdate = true
		} else {
			for _, env := range envs {
				if strings.Contains(rule.PathRegex, env) {
					shouldUpdate = true
					break
				}
			}
		}

		if shouldUpdate {
			if rotateKeyType == "pgp" {
				existingKeys := rule.GetPGPKeys()
				existingKeys = append(existingKeys, newKey)
				rule.PGP = existingKeys
			} else {
				existingKeys := rule.GetAgeKeys()
				existingKeys = append(existingKeys, newKey)
				rule.Age = existingKeys
			}
			rulesUpdated++
			fmt.Printf("✓ Added new key to rule: %s\n", rule.PathRegex)
		}
	}

	// Save intermediate .sops.yaml
	if err := cfg.Save(".sops.yaml"); err != nil {
		return fmt.Errorf("failed to save .sops.yaml: %w", err)
	}

	// Step 4: Re-encrypt all files with both old and new keys
	fmt.Println("\n🔒 Re-encrypting files with both old and new keys...")
	files, err := sops.FindSopsFiles(".")
	if err != nil {
		return fmt.Errorf("failed to find encrypted files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("⚠️  No encrypted files found")
	} else {
		var failed []string
		for i, file := range files {
			fmt.Printf("[%d/%d] %s...\n", i+1, len(files), file)
			if err := sops.Updatekeys(file); err != nil {
				fmt.Printf("  ❌ Failed: %s\n", err)
				failed = append(failed, file)
			} else {
				fmt.Println("  ✓ Updated")
			}
		}

		if len(failed) > 0 {
			return fmt.Errorf("failed to re-encrypt %d files", len(failed))
		}
	}

	// Step 5: Remove old keys
	fmt.Println("\n🧹 Removing old keys from .sops.yaml...")
	for i := range cfg.CreationRules {
		rule := &cfg.CreationRules[i]

		shouldUpdate := false
		if rotateEnvironments == "all" {
			shouldUpdate = true
		} else {
			for _, env := range envs {
				if strings.Contains(rule.PathRegex, env) {
					shouldUpdate = true
					break
				}
			}
		}

		if shouldUpdate {
			if rotateKeyType == "pgp" {
				existingKeys := rule.GetPGPKeys()
				newKeys := []string{}
				for _, key := range existingKeys {
					isOld := false
					for _, oldKey := range oldKeys {
						if key == oldKey {
							isOld = true
							break
						}
					}
					if !isOld {
						newKeys = append(newKeys, key)
					}
				}
				rule.PGP = newKeys
			} else {
				existingKeys := rule.GetAgeKeys()
				newKeys := []string{}
				for _, key := range existingKeys {
					isOld := false
					for _, oldKey := range oldKeys {
						if key == oldKey {
							isOld = true
							break
						}
					}
					if !isOld {
						newKeys = append(newKeys, key)
					}
				}
				rule.Age = newKeys
			}
			fmt.Printf("✓ Removed old keys from rule: %s\n", rule.PathRegex)
		}
	}

	// Save final .sops.yaml
	if err := cfg.Save(".sops.yaml"); err != nil {
		return fmt.Errorf("failed to save .sops.yaml: %w", err)
	}

	// Step 6: Final re-encryption with only new keys
	if len(files) > 0 {
		fmt.Println("\n🔒 Final re-encryption with new keys only...")
		var failed []string
		for i, file := range files {
			fmt.Printf("[%d/%d] %s...\n", i+1, len(files), file)
			if err := sops.Updatekeys(file); err != nil {
				fmt.Printf("  ❌ Failed: %s\n", err)
				failed = append(failed, file)
			} else {
				fmt.Println("  ✓ Updated")
			}
		}

		if len(failed) > 0 {
			return fmt.Errorf("failed to re-encrypt %d files", len(failed))
		}
	}

	fmt.Println("\n✅ Key rotation complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Verify files can still be decrypted: keyforge edit <file>")
	fmt.Println("  2. Commit .sops.yaml to git")
	if rotateKeyType == "pgp" {
		fmt.Printf("  3. Securely backup old key: gpg --export-secret-keys %s > backup.gpg\n", oldKeys[0])
		fmt.Println("  4. Consider removing old key from keyring after verification")
	} else {
		fmt.Println("  3. Securely backup old Age key from ~/.config/sops/age/keys.txt")
	}

	return nil
}
