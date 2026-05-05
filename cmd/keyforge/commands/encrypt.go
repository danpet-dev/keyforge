package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/danpet-dev/keyforge/pkg/sops"
	"github.com/spf13/cobra"
)

var (
	encryptOutput  string
	encryptInPlace bool
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt <file>",
	Short: "Encrypt file with SOPS",
	Long: `Encrypt a plaintext file with SOPS using .sops.yaml configuration.

The encrypted file will have .sops extension added automatically.

Examples:
  # Encrypt and create .sops file
  keyforge encrypt secrets/prod.yaml

  # Encrypt with custom output
  keyforge encrypt secrets/prod.yaml --output secrets/encrypted.yaml.sops

  # Encrypt in place (replace original with encrypted version)
  keyforge encrypt secrets/prod.yaml --in-place

Requirements:
  - .sops.yaml must exist in current or parent directory
  - File path must match creation_rules in .sops.yaml
  - Encryption keys (PGP/Age) must be configured

Security Note:
  After encryption, remember to:
  - Verify the encrypted file was created
  - Delete the plaintext file (unless using --in-place)
  - Add plaintext files to .gitignore
  - Commit only the .sops files to git`,
	Args: cobra.ExactArgs(1),
	RunE: runEncrypt,
}

func init() {
	rootCmd.AddCommand(encryptCmd)

	encryptCmd.Flags().StringVarP(&encryptOutput, "output", "o", "", "Output file (default: <file>.sops)")
	encryptCmd.Flags().BoolVarP(&encryptInPlace, "in-place", "i", false, "Encrypt in place (replace original)")
}

func runEncrypt(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Check if SOPS is installed
	if !sops.IsSopsInstalled() {
		return fmt.Errorf("SOPS is not installed. Please install it first:\n  https://github.com/getsops/sops")
	}

	// Check if file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", inputFile)
	}

	// Check if .sops.yaml exists
	if _, err := os.Stat(".sops.yaml"); os.IsNotExist(err) {
		// Try parent directories
		found := false
		for i := 0; i < 5; i++ {
			prefix := strings.Repeat("../", i+1)
			if _, err := os.Stat(prefix + ".sops.yaml"); err == nil {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf(".sops.yaml not found. Run 'keyforge init' to create one")
		}
	}

	// Determine output file
	outputFile := encryptOutput
	if outputFile == "" {
		outputFile = inputFile + ".sops"
	}

	// Check if output file already exists
	if _, err := os.Stat(outputFile); err == nil && !encryptInPlace {
		return fmt.Errorf("output file already exists: %s (use --in-place to overwrite)", outputFile)
	}

	// Encrypt file
	if err := sops.EncryptFile(inputFile, outputFile); err != nil {
		return fmt.Errorf("failed to encrypt file: %w", err)
	}

	fmt.Printf("✅ Encrypted: %s → %s\n", inputFile, outputFile)

	// In-place: delete original
	if encryptInPlace {
		if err := os.Remove(inputFile); err != nil {
			fmt.Printf("⚠️  Warning: Could not delete original file: %v\n", err)
		} else {
			fmt.Printf("🗑️  Deleted: %s\n", inputFile)
		}
	} else {
		// Suggest deleting original
		fmt.Printf("\n💡 Next steps:\n")
		fmt.Printf("   - Verify encrypted file: keyforge decrypt %s\n", outputFile)
		fmt.Printf("   - Delete plaintext: rm %s\n", inputFile)
		fmt.Printf("   - Add to .gitignore: echo '%s' >> .gitignore\n", inputFile)
		fmt.Printf("   - Commit encrypted: git add %s && git commit\n", outputFile)
	}

	return nil
}
