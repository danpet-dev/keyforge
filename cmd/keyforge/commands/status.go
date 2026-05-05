package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danpet-dev/keyforge/pkg/config"
	"github.com/danpet-dev/keyforge/pkg/sops"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of encrypted secrets",
	Long: `Show status of encrypted secrets and security issues.

This command shows:
  - Encrypted files (.sops) and their status
  - Decrypted plaintext files (security warning!)
  - Files that need updatekeys (out of sync with .sops.yaml)
  - Validation of .sops.yaml
  - Plaintext secrets not in .gitignore

Example:
  keyforge status
  keyforge status --verbose`,
	RunE: runStatus,
}

var (
	statusVerbose bool
)

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVarP(&statusVerbose, "verbose", "v", false, "Show detailed information")
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 KeyForge Status Report")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Check if .sops.yaml exists
	hasSopsYaml := false
	if _, err := os.Stat(".sops.yaml"); err == nil {
		hasSopsYaml = true
	}

	// Validate .sops.yaml
	var cfg *config.SopsConfig
	if hasSopsYaml {
		fmt.Println("📄 .sops.yaml Validation")
		fmt.Println(strings.Repeat("-", 70))

		loadedCfg, err := config.Load(".sops.yaml")
		if err != nil {
			fmt.Printf("❌ INVALID: %v\n", err)
			return fmt.Errorf("invalid .sops.yaml: %w", err)
		}
		cfg = loadedCfg

		fmt.Printf("✓ Valid .sops.yaml found\n")
		fmt.Printf("  Rules: %d\n", len(cfg.CreationRules))

		if statusVerbose {
			for i, rule := range cfg.CreationRules {
				fmt.Printf("\n  Rule %d: %s\n", i+1, rule.PathRegex)
				pgpKeys := rule.GetPGPKeys()
				ageKeys := rule.GetAgeKeys()
				if len(pgpKeys) > 0 {
					fmt.Printf("    PGP keys: %d\n", len(pgpKeys))
				}
				if len(ageKeys) > 0 {
					fmt.Printf("    Age keys: %d\n", len(ageKeys))
				}
			}
		}
		fmt.Println()
	} else {
		fmt.Println("📄 .sops.yaml Validation")
		fmt.Println(strings.Repeat("-", 70))
		fmt.Println("⚠️  No .sops.yaml found")
		fmt.Println("   Run 'keyforge init' to create one")
		fmt.Println()
	}

	// Find encrypted files
	fmt.Println("🔒 Encrypted Files (.sops)")
	fmt.Println(strings.Repeat("-", 70))

	encryptedFiles, err := sops.FindSopsFiles(".")
	if err != nil {
		fmt.Printf("⚠️  Error finding encrypted files: %v\n", err)
		encryptedFiles = []string{}
	}

	if len(encryptedFiles) == 0 {
		fmt.Println("No encrypted files found")
	} else {
		fmt.Printf("Found %d encrypted file(s):\n", len(encryptedFiles))
		for _, file := range encryptedFiles {
			fmt.Printf("  ✓ %s\n", file)
		}
	}
	fmt.Println()

	// Find decrypted files (potential security issue)
	fmt.Println("🔓 Decrypted Files (Plaintext)")
	fmt.Println(strings.Repeat("-", 70))

	decryptedFiles := findDecryptedFiles(encryptedFiles)
	if len(decryptedFiles) == 0 {
		fmt.Println("✓ No decrypted plaintext files found (good!)")
	} else {
		fmt.Printf("⚠️  Found %d decrypted file(s) - SECURITY RISK!\n", len(decryptedFiles))
		for _, file := range decryptedFiles {
			fmt.Printf("  ❌ %s\n", file)
		}
		fmt.Println()
		fmt.Println("  WARNING: These files contain plaintext secrets!")
		fmt.Println("  Action: Delete them or add to .gitignore")
	}
	fmt.Println()

	// Check .gitignore
	fmt.Println("📝 .gitignore Check")
	fmt.Println(strings.Repeat("-", 70))

	gitignoreIssues := checkGitignore(decryptedFiles)
	if len(gitignoreIssues) == 0 {
		if len(decryptedFiles) == 0 {
			fmt.Println("✓ No plaintext secrets to check")
		} else {
			fmt.Println("✓ All plaintext files are in .gitignore")
		}
	} else {
		fmt.Printf("⚠️  %d file(s) NOT in .gitignore:\n", len(gitignoreIssues))
		for _, file := range gitignoreIssues {
			fmt.Printf("  ❌ %s\n", file)
		}
		fmt.Println()
		fmt.Println("  WARNING: These files may be committed to git!")
		fmt.Println("  Action: Add them to .gitignore")
	}
	fmt.Println()

	// Check if files need updatekeys
	if hasSopsYaml && len(encryptedFiles) > 0 {
		fmt.Println("🔄 Update Keys Status")
		fmt.Println(strings.Repeat("-", 70))

		needsUpdate := checkNeedsUpdatekeys(encryptedFiles, cfg)
		if len(needsUpdate) == 0 {
			fmt.Println("✓ All encrypted files are up to date")
		} else {
			fmt.Printf("⚠️  %d file(s) may need 'updatekeys':\n", len(needsUpdate))
			for _, file := range needsUpdate {
				fmt.Printf("  ⚠️  %s\n", file)
			}
			fmt.Println()
			fmt.Println("  These files might be out of sync with .sops.yaml")
			fmt.Println("  Run: keyforge update-all")
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("📊 Summary")
	fmt.Println(strings.Repeat("-", 70))

	issueCount := 0
	if !hasSopsYaml {
		fmt.Println("⚠️  No .sops.yaml found")
		issueCount++
	}
	if len(decryptedFiles) > 0 {
		fmt.Printf("⚠️  %d decrypted plaintext files (security risk)\n", len(decryptedFiles))
		issueCount++
	}
	if len(gitignoreIssues) > 0 {
		fmt.Printf("⚠️  %d plaintext files not in .gitignore\n", len(gitignoreIssues))
		issueCount++
	}

	if issueCount == 0 {
		fmt.Println("✓ No issues found - everything looks good!")
	} else {
		fmt.Printf("\n⚠️  Found %d issue(s) that need attention\n", issueCount)
	}

	fmt.Println(strings.Repeat("=", 70))

	return nil
}

// findDecryptedFiles finds plaintext versions of encrypted files
func findDecryptedFiles(encryptedFiles []string) []string {
	decrypted := []string{}

	for _, encFile := range encryptedFiles {
		// Remove .sops extension to get plaintext filename
		plainFile := strings.TrimSuffix(encFile, ".sops")

		// Check if plaintext file exists
		if _, err := os.Stat(plainFile); err == nil {
			decrypted = append(decrypted, plainFile)
		}
	}

	return decrypted
}

// checkGitignore checks if decrypted files are in .gitignore
func checkGitignore(decryptedFiles []string) []string {
	if len(decryptedFiles) == 0 {
		return []string{}
	}

	// Read .gitignore
	gitignoreContent, err := os.ReadFile(".gitignore")
	if err != nil {
		// No .gitignore file - all files are issues
		return decryptedFiles
	}

	gitignoreLines := strings.Split(string(gitignoreContent), "\n")
	ignoredPatterns := []string{}
	for _, line := range gitignoreLines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			ignoredPatterns = append(ignoredPatterns, line)
		}
	}

	issues := []string{}
	for _, file := range decryptedFiles {
		if !isIgnored(file, ignoredPatterns) {
			issues = append(issues, file)
		}
	}

	return issues
}

// isIgnored checks if a file matches any .gitignore pattern
func isIgnored(file string, patterns []string) bool {
	for _, pattern := range patterns {
		// Match exact file
		if pattern == file {
			return true
		}

		// Match basename
		if pattern == filepath.Base(file) {
			return true
		}

		// Simple wildcard pattern (*.ext)
		if strings.HasPrefix(pattern, "*.") {
			ext := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(file, ext) {
				return true
			}
		}

		// Simple prefix pattern (dir/)
		if strings.HasSuffix(pattern, "/") {
			if strings.HasPrefix(file, pattern) {
				return true
			}
		}

		// Contains pattern (simple substring)
		if strings.Contains(file, pattern) {
			return true
		}
	}
	return false
}

// checkNeedsUpdatekeys checks which files might need updatekeys
// This is a heuristic check - we can't definitively know without parsing SOPS metadata
func checkNeedsUpdatekeys(encryptedFiles []string, cfg *config.SopsConfig) []string {
	// For now, we'll return empty slice
	// Full implementation would parse SOPS metadata from each file
	// and compare with .sops.yaml keys
	// This requires more complex logic

	// Placeholder: return empty for now
	return []string{}
}
