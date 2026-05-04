package commands

import (
	"fmt"
	"os"

	"github.com/danpet-dev/keyforge/pkg/config"
	"github.com/danpet-dev/keyforge/pkg/keys"
	"github.com/danpet-dev/keyforge/pkg/sops"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate .sops.yaml configuration",
	Long: `Validate .sops.yaml syntax and check key availability.

This command checks:
  - .sops.yaml syntax is valid YAML
  - All referenced PGP keys are available in keyring
  - Keys are not expired or expiring soon
  - SOPS is installed and accessible

Example:
  keyforge validate`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Check if .sops.yaml exists
	if _, err := os.Stat(".sops.yaml"); os.IsNotExist(err) {
		return fmt.Errorf(".sops.yaml not found. Run 'keyforge init' to create one")
	}

	// Check if SOPS is installed
	if !sops.IsSopsInstalled() {
		return fmt.Errorf("SOPS is not installed or not in PATH")
	}

	version, err := sops.GetSopsVersion()
	if err == nil {
		fmt.Printf("✓ SOPS installed: %s\n", version)
	}

	// Load .sops.yaml
	cfg, err := config.Load(".sops.yaml")
	if err != nil {
		return fmt.Errorf("failed to load .sops.yaml: %w", err)
	}

	fmt.Printf("✓ .sops.yaml syntax is valid\n")
	fmt.Printf("✓ Found %d creation rules\n", len(cfg.CreationRules))

	var missingKeys []string
	var expiringKeys []string

	// Check PGP keys
	if keys.IsGPGInstalled() {
		keyList, err := keys.ListPGPKeys()
		if err != nil {
			return fmt.Errorf("failed to list PGP keys: %w", err)
		}

		fmt.Printf("\n🔑 PGP Keys in keyring: %d\n", len(keyList))

		// Build map of available keys
		availableKeys := make(map[string]keys.Key)
		for _, key := range keyList {
			availableKeys[key.Fingerprint] = key
		}

		// Check each rule
		for i, rule := range cfg.CreationRules {
			pgpKeys := rule.GetPGPKeys()
			if len(pgpKeys) > 0 {
				fmt.Printf("\nRule %d: %s (PGP)\n", i+1, rule.PathRegex)

				for _, fp := range pgpKeys {
					key, exists := availableKeys[fp]
					if !exists {
						missingKeys = append(missingKeys, fp)
						fmt.Printf("  ❌ Missing key: %s\n", fp)
						continue
					}

					fmt.Printf("  ✓ %s (%s)\n", fp, key.Email)

					// Check expiration
					if expired, msg := keys.CheckKeyExpiration(key); expired {
						expiringKeys = append(expiringKeys, fmt.Sprintf("%s: %s", fp, msg))
						fmt.Printf("    ⚠️  %s\n", msg)
					}
				}
			}
		}
	} else {
		fmt.Println("⚠️  GPG not installed, skipping PGP key validation")
	}

	// Check Age keys
	ageKeyList, err := keys.ListAgeKeys()
	if err == nil && len(ageKeyList) > 0 {
		fmt.Printf("\n🔑 Age Keys available: %d\n", len(ageKeyList))

		// Build map of available age keys
		availableAgeKeys := make(map[string]keys.Key)
		for _, key := range ageKeyList {
			availableAgeKeys[key.PublicKey] = key
		}

		// Check each rule
		for i, rule := range cfg.CreationRules {
			ageKeys := rule.GetAgeKeys()
			if len(ageKeys) > 0 {
				fmt.Printf("\nRule %d: %s (Age)\n", i+1, rule.PathRegex)

				for _, pubKey := range ageKeys {
					key, exists := availableAgeKeys[pubKey]
					if !exists {
						missingKeys = append(missingKeys, pubKey)
						fmt.Printf("  ❌ Missing key: %s\n", pubKey)
						continue
					}

					name := key.Name
					if name == "" {
						name = "unnamed"
					}
					fmt.Printf("  ✓ %s (%s)\n", pubKey, name)
				}
			}
		}
	}

	// Summary
	fmt.Println("\n📋 Summary:")
	if len(missingKeys) > 0 {
		fmt.Printf("  ❌ %d missing keys\n", len(missingKeys))
		return fmt.Errorf("validation failed: missing keys")
	}

	if len(expiringKeys) > 0 {
		fmt.Printf("  ⚠️  %d expiring/expired keys\n", len(expiringKeys))
	} else {
		fmt.Println("  ✓ All keys valid and available")
	}

	// Find encrypted files
	files, err := sops.FindSopsFiles(".")
	if err == nil && len(files) > 0 {
		fmt.Printf("\n📁 Found %d encrypted files:\n", len(files))
		for _, file := range files {
			fmt.Printf("  - %s\n", file)
		}
	}

	return nil
}
