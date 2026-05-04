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

var addMemberCmd = &cobra.Command{
	Use:   "add-member",
	Short: "Add a team member to .sops.yaml",
	Long: `Add a new team member's encryption key to .sops.yaml.

This command:
  1. Optionally generates a new PGP key for the member
  2. Adds the key to specified environments in .sops.yaml
  3. Runs 'updatekeys' on all affected encrypted files
  4. Exports the public key for sharing

Example:
  keyforge add-member --name "Alice" --email alice@example.com --environments dev,test
  keyforge add-member --name "Bob" --email bob@example.com --generate-key`,
	RunE: runAddMember,
}

var (
	addMemberName         string
	addMemberEmail        string
	addMemberKeyType      string
	addMemberEnvironments string
	addMemberGenerateKey  bool
	addMemberFingerprint  string
	addMemberPublicKey    string
)

func init() {
	rootCmd.AddCommand(addMemberCmd)

	addMemberCmd.Flags().StringVarP(&addMemberName, "name", "n", "", "Member name (required)")
	addMemberCmd.Flags().StringVarP(&addMemberEmail, "email", "e", "", "Member email (required)")
	addMemberCmd.Flags().StringVarP(&addMemberKeyType, "key-type", "t", "pgp", "Key type: pgp or age")
	addMemberCmd.Flags().StringVarP(&addMemberEnvironments, "environments", "E", "all", "Environments (comma-separated: dev,test,prod or 'all')")
	addMemberCmd.Flags().BoolVarP(&addMemberGenerateKey, "generate-key", "g", false, "Generate a new PGP or Age key")
	addMemberCmd.Flags().StringVarP(&addMemberFingerprint, "fingerprint", "f", "", "Use existing PGP key fingerprint")
	addMemberCmd.Flags().StringVarP(&addMemberPublicKey, "public-key", "p", "", "Use existing Age public key")

	if err := addMemberCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
}

func runAddMember(cmd *cobra.Command, args []string) error {
	// Check if .sops.yaml exists
	if _, err := os.Stat(".sops.yaml"); os.IsNotExist(err) {
		return fmt.Errorf(".sops.yaml not found. Run 'keyforge init' first")
	}

	// Validate key type
	if addMemberKeyType != "pgp" && addMemberKeyType != "age" {
		return fmt.Errorf("invalid key-type: %s (must be 'pgp' or 'age')", addMemberKeyType)
	}

	// Check prerequisites
	if addMemberKeyType == "pgp" && !keys.IsGPGInstalled() {
		return fmt.Errorf("GPG is not installed. Please install GPG first")
	}

	var keyValue string

	// Generate or use existing key
	if addMemberGenerateKey {
		if addMemberKeyType == "pgp" {
			if addMemberEmail == "" {
				return fmt.Errorf("--email is required for PGP key generation")
			}
			fmt.Printf("Generating PGP key for %s (%s)...\n", addMemberName, addMemberEmail)
			fp, err := keys.GeneratePGPKey(addMemberName, addMemberEmail, 2)
			if err != nil {
				return fmt.Errorf("failed to generate PGP key: %w", err)
			}
			keyValue = fp
			fmt.Printf("✓ Generated PGP key: %s\n", keyValue)
		} else if addMemberKeyType == "age" {
			fmt.Printf("Generating Age key for %s...\n", addMemberName)
			identity, err := keys.GenerateAgeKey()
			if err != nil {
				return fmt.Errorf("failed to generate Age key: %w", err)
			}

			comment := fmt.Sprintf("%s - created by keyforge", addMemberName)
			if err := keys.SaveAgeKey(identity, comment); err != nil {
				return fmt.Errorf("failed to save Age key: %w", err)
			}

			keyValue = identity.Recipient().String()
			fmt.Printf("✓ Generated Age key: %s\n", keyValue)
			fmt.Printf("✓ Saved to ~/.config/sops/age/keys.txt\n")
		}
	} else if addMemberFingerprint != "" && addMemberKeyType == "pgp" {
		keyValue = addMemberFingerprint
		fmt.Printf("Using existing PGP key: %s\n", keyValue)
	} else if addMemberPublicKey != "" && addMemberKeyType == "age" {
		keyValue = addMemberPublicKey
		fmt.Printf("Using existing Age key: %s\n", keyValue)
	} else {
		// Search for key
		if addMemberKeyType == "pgp" {
			if addMemberEmail == "" {
				return fmt.Errorf("--email is required for PGP keys")
			}
			keyList, err := keys.ListPGPKeys()
			if err != nil {
				return fmt.Errorf("failed to list PGP keys: %w", err)
			}

			found := false
			for _, key := range keyList {
				if key.Email == addMemberEmail {
					keyValue = key.Fingerprint
					found = true
					fmt.Printf("Found existing key: %s (%s)\n", keyValue, addMemberEmail)
					break
				}
			}

			if !found {
				return fmt.Errorf("no key found for %s. Use --generate-key or --fingerprint", addMemberEmail)
			}
		} else if addMemberKeyType == "age" {
			return fmt.Errorf("for Age keys, use --public-key or --generate-key")
		}
	}

	// Load .sops.yaml
	cfg, err := config.Load(".sops.yaml")
	if err != nil {
		return fmt.Errorf("failed to load .sops.yaml: %w", err)
	}

	// Parse environments
	envs := strings.Split(addMemberEnvironments, ",")
	if addMemberEnvironments == "all" {
		envs = []string{"all"}
	}

	// Add key to specified environments
	rulesUpdated := 0
	for i := range cfg.CreationRules {
		rule := &cfg.CreationRules[i]

		// Check if this rule matches requested environments
		shouldUpdate := false
		if addMemberEnvironments == "all" {
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
			// Check if key already exists
			var existingKeys []string
			var keyExists bool

			if addMemberKeyType == "pgp" {
				existingKeys = rule.GetPGPKeys()
			} else {
				existingKeys = rule.GetAgeKeys()
			}

			for _, existingKey := range existingKeys {
				if existingKey == keyValue {
					keyExists = true
					break
				}
			}

			if !keyExists {
				existingKeys = append(existingKeys, keyValue)
				if addMemberKeyType == "pgp" {
					rule.PGP = existingKeys
				} else {
					rule.Age = existingKeys
				}
				rulesUpdated++
				fmt.Printf("✓ Added key to rule: %s\n", rule.PathRegex)
			} else {
				fmt.Printf("⚠️  Key already exists in rule: %s\n", rule.PathRegex)
			}
		}
	}

	if rulesUpdated == 0 {
		return fmt.Errorf("no rules were updated. Check environment names or use --environments all")
	}

	// Save .sops.yaml
	if err := cfg.Save(".sops.yaml"); err != nil {
		return fmt.Errorf("failed to save .sops.yaml: %w", err)
	}

	fmt.Printf("✓ Updated .sops.yaml (%d rules)\n", rulesUpdated)

	// Export public key (PGP only)
	if addMemberKeyType == "pgp" {
		publicKey, err := keys.ExportPublicKey(keyValue)
		if err != nil {
			fmt.Printf("⚠️  Failed to export public key: %s\n", err)
		} else {
			keyFile := fmt.Sprintf("%s.asc", addMemberEmail)
			if err := os.WriteFile(keyFile, []byte(publicKey), 0644); err != nil {
				fmt.Printf("⚠️  Failed to write key file: %s\n", err)
			} else {
				fmt.Printf("✓ Exported public key: %s\n", keyFile)
			}
		}
	} else if addMemberKeyType == "age" {
		fmt.Printf("✓ Age public key: %s\n", keyValue)
		fmt.Println("  Share this public key with team members who need to add you to their .sops.yaml")
	}

	// Update encrypted files
	fmt.Println("\nUpdating encrypted files...")
	files, err := sops.FindSopsFiles(".")
	if err != nil {
		return fmt.Errorf("failed to find encrypted files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No encrypted files found")
		return nil
	}

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
		fmt.Printf("\n❌ Failed to update %d files\n", len(failed))
		return fmt.Errorf("some files failed to update")
	}

	fmt.Println("\n✓ All files updated successfully")
	fmt.Println("\nNext steps:")
	if addMemberKeyType == "pgp" && addMemberEmail != "" {
		fmt.Printf("  1. Share public key with %s: %s.asc\n", addMemberName, addMemberEmail)
		fmt.Printf("  2. %s imports key: gpg --import %s.asc\n", addMemberName, addMemberEmail)
		fmt.Println("  3. Commit .sops.yaml to git")
	} else {
		fmt.Printf("  1. Share Age public key with %s: %s\n", addMemberName, keyValue)
		fmt.Println("  2. Commit .sops.yaml to git")
	}

	return nil
}
