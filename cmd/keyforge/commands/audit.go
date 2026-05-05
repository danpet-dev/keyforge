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

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit key access and permissions",
	Long: `Show detailed audit of who has access to which secrets.

This command provides:
  - List of all keys in .sops.yaml
  - Which environments each key can decrypt
  - Owner information (name, email) for each key
  - Expiration status for PGP keys
  - List of encrypted files matching each rule

Example:
  keyforge audit
  keyforge audit --format json
  keyforge audit --key B016BDB3E04C2D186389850787070D91C80C0400`,
	RunE: runAudit,
}

var (
	auditFormat string
	auditKey    string
)

func init() {
	rootCmd.AddCommand(auditCmd)

	auditCmd.Flags().StringVarP(&auditFormat, "format", "f", "text", "Output format: text or json")
	auditCmd.Flags().StringVarP(&auditKey, "key", "k", "", "Show access for specific key only")
}

func runAudit(cmd *cobra.Command, args []string) error {
	// Check if .sops.yaml exists
	if _, err := os.Stat(".sops.yaml"); os.IsNotExist(err) {
		return fmt.Errorf(".sops.yaml not found. Run 'keyforge init' first")
	}

	// Load .sops.yaml
	cfg, err := config.Load(".sops.yaml")
	if err != nil {
		return fmt.Errorf("failed to load .sops.yaml: %w", err)
	}

	// Load available keys
	var pgpKeys map[string]keys.Key
	var ageKeys map[string]keys.Key

	if keys.IsGPGInstalled() {
		keyList, err := keys.ListPGPKeys()
		if err == nil {
			pgpKeys = make(map[string]keys.Key)
			for _, key := range keyList {
				pgpKeys[key.Fingerprint] = key
			}
		}
	}

	ageKeyList, err := keys.ListAgeKeys()
	if err == nil {
		ageKeys = make(map[string]keys.Key)
		for _, key := range ageKeyList {
			ageKeys[key.PublicKey] = key
		}
	}

	// Find encrypted files
	files, err := sops.FindSopsFiles(".")
	if err != nil {
		files = []string{}
	}

	// Build access map: key -> rules
	type KeyAccess struct {
		Key         string
		Type        string
		Name        string
		Email       string
		Expires     string
		Available   bool
		Rules       []string
		FilesAccess []string
	}

	accessMap := make(map[string]*KeyAccess)

	// Analyze each rule
	for _, rule := range cfg.CreationRules {
		// PGP keys
		for _, fingerprint := range rule.GetPGPKeys() {
			if auditKey != "" && fingerprint != auditKey {
				continue
			}

			if _, exists := accessMap[fingerprint]; !exists {
				access := &KeyAccess{
					Key:   fingerprint,
					Type:  "PGP",
					Rules: []string{},
				}

				if key, found := pgpKeys[fingerprint]; found {
					access.Name = key.Name
					access.Email = key.Email
					access.Available = true
					if key.Expires != nil {
						access.Expires = key.Expires.Format("2006-01-02")
					} else {
						access.Expires = "never"
					}
				} else {
					access.Available = false
					access.Expires = "unknown"
				}

				accessMap[fingerprint] = access
			}

			accessMap[fingerprint].Rules = append(accessMap[fingerprint].Rules, rule.PathRegex)
		}

		// Age keys
		for _, pubKey := range rule.GetAgeKeys() {
			if auditKey != "" && pubKey != auditKey {
				continue
			}

			if _, exists := accessMap[pubKey]; !exists {
				access := &KeyAccess{
					Key:     pubKey,
					Type:    "Age",
					Rules:   []string{},
					Expires: "never",
				}

				if key, found := ageKeys[pubKey]; found {
					access.Name = key.Name
					access.Available = true
				} else {
					access.Available = false
				}

				accessMap[pubKey] = access
			}

			accessMap[pubKey].Rules = append(accessMap[pubKey].Rules, rule.PathRegex)
		}
	}

	// Match files to keys
	for _, access := range accessMap {
		for _, file := range files {
			for _, rule := range access.Rules {
				matched, _ := matchPathRegex(file, rule)
				if matched {
					access.FilesAccess = append(access.FilesAccess, file)
					break
				}
			}
		}
	}

	// Output
	if auditFormat == "json" {
		// JSON output (simplified for now)
		fmt.Println("{")
		fmt.Println("  \"keys\": [")
		first := true
		for _, access := range accessMap {
			if !first {
				fmt.Println(",")
			}
			first = false
			fmt.Printf("    {\n")
			fmt.Printf("      \"key\": \"%s\",\n", access.Key)
			fmt.Printf("      \"type\": \"%s\",\n", access.Type)
			fmt.Printf("      \"name\": \"%s\",\n", access.Name)
			fmt.Printf("      \"email\": \"%s\",\n", access.Email)
			fmt.Printf("      \"expires\": \"%s\",\n", access.Expires)
			fmt.Printf("      \"available\": %v,\n", access.Available)
			fmt.Printf("      \"rules_count\": %d,\n", len(access.Rules))
			fmt.Printf("      \"files_access\": %d\n", len(access.FilesAccess))
			fmt.Printf("    }")
		}
		fmt.Println("\n  ]")
		fmt.Println("}")
	} else {
		// Text output
		fmt.Println("🔍 KeyForge Audit Report")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("\nTotal keys: %d\n", len(accessMap))
		fmt.Printf("Total rules: %d\n", len(cfg.CreationRules))
		fmt.Printf("Total encrypted files: %d\n\n", len(files))

		for keyID, access := range accessMap {
			fmt.Println(strings.Repeat("-", 60))
			if access.Type == "PGP" {
				fmt.Printf("🔑 PGP Key: %s\n", keyID[:16]+"...")
				if access.Name != "" {
					fmt.Printf("   Owner: %s <%s>\n", access.Name, access.Email)
				}
			} else {
				fmt.Printf("🔑 Age Key: %s\n", keyID[:16]+"...")
				if access.Name != "" {
					fmt.Printf("   Owner: %s\n", access.Name)
				}
			}

			if access.Available {
				fmt.Println("   Status: ✓ Available")
			} else {
				fmt.Println("   Status: ❌ Not available (cannot decrypt)")
			}

			if access.Type == "PGP" {
				fmt.Printf("   Expires: %s\n", access.Expires)
			}

			fmt.Printf("\n   Access to %d rule(s):\n", len(access.Rules))
			for _, rule := range access.Rules {
				fmt.Printf("     - %s\n", rule)
			}

			if len(access.FilesAccess) > 0 {
				fmt.Printf("\n   Can decrypt %d file(s):\n", len(access.FilesAccess))
				for _, file := range access.FilesAccess {
					fmt.Printf("     - %s\n", file)
				}
			} else {
				fmt.Println("\n   No encrypted files match these rules")
			}

			fmt.Println()
		}

		fmt.Println(strings.Repeat("=", 60))

		// Security warnings
		hasWarnings := false
		for _, access := range accessMap {
			if !access.Available {
				if !hasWarnings {
					fmt.Println("\n⚠️  Security Warnings:")
					hasWarnings = true
				}
				fmt.Printf("  - Key %s is not available but still in .sops.yaml\n", access.Key[:16]+"...")
			}
		}

		if hasWarnings {
			fmt.Println("\nRecommendation: Remove unavailable keys or import them to your keyring")
		}
	}

	return nil
}

// matchPathRegex checks if a file path matches a regex pattern
func matchPathRegex(path, regex string) (bool, error) {
	// Simple string matching for now (proper regex would use regexp package)
	// This handles basic cases like "secrets/development\.yaml\.sops$"

	// Remove anchors
	pattern := strings.TrimPrefix(regex, "^")
	pattern = strings.TrimSuffix(pattern, "$")

	// Unescape dots
	pattern = strings.ReplaceAll(pattern, "\\.", ".")

	// Check if path ends with pattern (for simple cases)
	if strings.HasSuffix(path, pattern) {
		return true, nil
	}

	// Check if pattern contains wildcards
	if strings.Contains(pattern, ".*") {
		// Very basic wildcard matching
		parts := strings.Split(pattern, ".*")
		allMatch := true
		for _, part := range parts {
			if part != "" && !strings.Contains(path, part) {
				allMatch = false
				break
			}
		}
		return allMatch, nil
	}

	return false, nil
}
