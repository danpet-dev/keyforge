package commands

import (
	"fmt"
	"os"

	"github.com/danpet-dev/keyforge/pkg/config"
	"github.com/danpet-dev/keyforge/pkg/keys"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .sops.yaml with best-practice configuration",
	Long: `Initialize a new .sops.yaml file with multi-environment setup.

This command creates a .sops.yaml file with preconfigured rules for:
  - Environment-specific secrets (development, test, prod)
  - Service-specific secrets (services/*/k8s/secrets/)
  - Database secrets (database/*/secrets/)

Example:
  keyforge init --project my-project
  keyforge init --project my-project --key-type age
  keyforge init --project my-project --generate-key`,
	RunE: runInit,
}

var (
	initProject     string
	initKeyType     string
	initGenerateKey bool
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initProject, "project", "p", "", "Project name (required)")
	initCmd.Flags().StringVarP(&initKeyType, "key-type", "t", "pgp", "Key type: pgp or age")
	initCmd.Flags().BoolVarP(&initGenerateKey, "generate-key", "g", false, "Generate a new PGP or Age key")

	initCmd.MarkFlagRequired("project")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if .sops.yaml already exists
	if _, err := os.Stat(".sops.yaml"); err == nil {
		return fmt.Errorf(".sops.yaml already exists, refusing to overwrite")
	}

	// Validate key type
	if initKeyType != "pgp" && initKeyType != "age" {
		return fmt.Errorf("invalid key-type: %s (must be 'pgp' or 'age')", initKeyType)
	}

	// Check prerequisites
	if initKeyType == "pgp" && !keys.IsGPGInstalled() {
		return fmt.Errorf("GPG is not installed. Please install GPG first")
	}

	var masterKey string

	// Generate key if requested
	if initGenerateKey {
		if initKeyType == "pgp" {
			email := fmt.Sprintf("admin@%s.local", initProject)
			name := fmt.Sprintf("%s Admin", initProject)

			fmt.Printf("Generating PGP key for %s (%s)...\n", name, email)
			fingerprint, err := keys.GeneratePGPKey(name, email, 2)
			if err != nil {
				return fmt.Errorf("failed to generate PGP key: %w", err)
			}

			fmt.Printf("✓ Generated PGP key: %s\n", fingerprint)
			masterKey = fingerprint
		} else if initKeyType == "age" {
			fmt.Printf("Generating Age key for %s...\n", initProject)
			identity, err := keys.GenerateAgeKey()
			if err != nil {
				return fmt.Errorf("failed to generate Age key: %w", err)
			}

			// Save to default location
			comment := fmt.Sprintf("%s master key - created by keyforge", initProject)
			if err := keys.SaveAgeKey(identity, comment); err != nil {
				return fmt.Errorf("failed to save Age key: %w", err)
			}

			masterKey = identity.Recipient().String()
			fmt.Printf("✓ Generated Age key: %s\n", masterKey)
			fmt.Printf("✓ Saved to ~/.config/sops/age/keys.txt\n")
		}
	} else {
		// Use existing keys
		if initKeyType == "pgp" {
			keyList, err := keys.ListPGPKeys()
			if err != nil {
				return fmt.Errorf("failed to list PGP keys: %w", err)
			}

			if len(keyList) == 0 {
				return fmt.Errorf("no PGP keys found. Use --generate-key or create a key manually with 'gpg --generate-key'")
			}

			// Use the first key as master key
			masterKey = keyList[0].Fingerprint
			fmt.Printf("Using existing PGP key: %s (%s)\n", keyList[0].Fingerprint, keyList[0].Email)
		} else if initKeyType == "age" {
			keyList, err := keys.ListAgeKeys()
			if err != nil {
				return fmt.Errorf("failed to list Age keys: %w", err)
			}

			if len(keyList) == 0 {
				return fmt.Errorf("no Age keys found. Use --generate-key or create a key manually")
			}

			// Use the first key as master key
			masterKey = keyList[0].PublicKey
			fmt.Printf("Using existing Age key: %s\n", keyList[0].PublicKey)
		}
	}

	// Generate .sops.yaml
	keyMap := map[string][]string{
		"development": {masterKey},
		"test":        {masterKey},
		"prod":        {masterKey},
		"all":         {masterKey},
	}

	cfg := config.GenerateTemplate(initProject, initKeyType, keyMap)

	if err := cfg.Save(".sops.yaml"); err != nil {
		return fmt.Errorf("failed to save .sops.yaml: %w", err)
	}

	fmt.Println("✓ Created .sops.yaml")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1. Review .sops.yaml and adjust paths if needed")
	fmt.Println("  2. Create secrets directory: mkdir -p secrets")
	fmt.Println("  3. Add team members: keyforge add-member --name \"Alice\" --email alice@example.com")
	fmt.Println("  4. Create encrypted file: sops secrets/development.yaml.sops")

	return nil
}
