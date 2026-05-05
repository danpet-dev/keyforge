package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danpet-dev/keyforge/pkg/sops"
	"github.com/spf13/cobra"
)

var (
	decryptOutput string
	decryptFormat string
	decryptInPlace bool
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt <file>",
	Short: "Decrypt SOPS-encrypted file",
	Long: `Decrypt a SOPS-encrypted file to stdout or file.

Examples:
  # Decrypt to stdout
  keyforge decrypt secrets/prod.yaml.sops

  # Decrypt to file
  keyforge decrypt secrets/prod.yaml.sops --output secrets/prod.yaml

  # Decrypt in place (remove .sops extension)
  keyforge decrypt secrets/prod.yaml.sops --in-place

  # Convert format during decryption
  keyforge decrypt secrets/prod.yaml.sops --format json --output secrets/prod.json

Supported formats: yaml, json, env, dotenv

Security Warning:
  Decrypted files contain plaintext secrets! Make sure to:
  - Add them to .gitignore
  - Delete them after use
  - Never commit them to version control`,
	Args: cobra.ExactArgs(1),
	RunE: runDecrypt,
}

func init() {
	rootCmd.AddCommand(decryptCmd)

	decryptCmd.Flags().StringVarP(&decryptOutput, "output", "o", "", "Output file (default: stdout)")
	decryptCmd.Flags().StringVarP(&decryptFormat, "format", "f", "", "Output format: yaml, json, env, dotenv (default: same as input)")
	decryptCmd.Flags().BoolVarP(&decryptInPlace, "in-place", "i", false, "Decrypt in place (remove .sops extension)")
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Check if SOPS is installed
	if !sops.IsSopsInstalled() {
		return fmt.Errorf("SOPS is not installed. Please install it first:\n  https://github.com/getsops/sops")
	}

	// Check if file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", inputFile)
	}

	// Determine output file
	outputFile := decryptOutput
	if decryptInPlace {
		if !strings.HasSuffix(inputFile, ".sops") {
			return fmt.Errorf("--in-place requires file with .sops extension")
		}
		outputFile = strings.TrimSuffix(inputFile, ".sops")
	}

	// Detect input format
	inputFormat := detectFormat(inputFile)

	// Determine output format
	format := decryptFormat
	if format == "" {
		if outputFile != "" {
			format = detectFormat(outputFile)
		} else {
			format = inputFormat
		}
	}

	// Decrypt file
	content, err := sops.Decrypt(inputFile)
	if err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	// Convert format if needed
	if format != inputFormat {
		converted, err := convertFormat(content, inputFormat, format)
		if err != nil {
			return fmt.Errorf("failed to convert format: %w", err)
		}
		content = converted
	}

	// Output
	if outputFile != "" {
		// Write to file
		if err := os.WriteFile(outputFile, content, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Printf("✅ Decrypted: %s → %s\n", inputFile, outputFile)
		
		// Security warning
		fmt.Printf("\n⚠️  Security Warning:\n")
		fmt.Printf("   %s contains plaintext secrets!\n", outputFile)
		fmt.Printf("   - Add to .gitignore\n")
		fmt.Printf("   - Delete after use\n")
		fmt.Printf("   - Never commit to git\n")

		// Check .gitignore
		if !isInGitignore(outputFile) {
			fmt.Printf("\n⚠️  File NOT in .gitignore! Add it:\n")
			fmt.Printf("   echo '%s' >> .gitignore\n", outputFile)
		}
	} else {
		// Write to stdout
		fmt.Print(string(content))
	}

	return nil
}

func detectFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Remove .sops extension if present
	if ext == ".sops" {
		filename = strings.TrimSuffix(filename, ext)
		ext = strings.ToLower(filepath.Ext(filename))
	}

	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".env":
		return "env"
	default:
		return "yaml" // default
	}
}

func convertFormat(content []byte, fromFormat, toFormat string) ([]byte, error) {
	// If same format, no conversion needed
	if fromFormat == toFormat {
		return content, nil
	}

	// For now, we'll delegate to SOPS for format conversion
	// SOPS can handle yaml<->json natively
	// For env/dotenv, we'll need custom conversion

	if (fromFormat == "yaml" && toFormat == "json") || (fromFormat == "json" && toFormat == "yaml") {
		// SOPS handles this during decrypt with --output-type flag
		// But we already decrypted, so we need to parse and convert
		// TODO: Implement YAML<->JSON conversion
		return nil, fmt.Errorf("format conversion not yet implemented (from %s to %s)", fromFormat, toFormat)
	}

	if toFormat == "env" || toFormat == "dotenv" {
		// Convert YAML/JSON to ENV format
		// TODO: Implement YAML/JSON -> ENV conversion
		return nil, fmt.Errorf("format conversion to %s not yet implemented", toFormat)
	}

	return nil, fmt.Errorf("unsupported format conversion: %s -> %s", fromFormat, toFormat)
}

func isInGitignore(file string) bool {
	// Read .gitignore
	content, err := os.ReadFile(".gitignore")
	if err != nil {
		return false
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Simple pattern matching
		if line == file || line == filepath.Base(file) {
			return true
		}

		// Wildcard patterns
		if strings.HasPrefix(line, "*.") {
			ext := strings.TrimPrefix(line, "*")
			if strings.HasSuffix(file, ext) {
				return true
			}
		}

		// Directory patterns
		if strings.HasSuffix(line, "/") && strings.HasPrefix(file, line) {
			return true
		}

		// Contains pattern
		if strings.Contains(file, line) {
			return true
		}
	}

	return false
}
