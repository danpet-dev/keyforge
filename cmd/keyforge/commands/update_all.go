package commands

import (
	"fmt"

	"github.com/danpet-dev/keyforge/pkg/sops"
	"github.com/spf13/cobra"
)

var updateAllCmd = &cobra.Command{
	Use:   "update-all",
	Short: "Update all encrypted files with new keys",
	Long: `Update all *.sops files with new encryption keys.

This command finds all encrypted files in the current directory (recursively)
and runs 'sops updatekeys' on each file with proper format detection.

Useful after adding new team members or rotating keys.

Example:
  keyforge update-all
  keyforge update-all --directory ./secrets`,
	RunE: runUpdateAll,
}

var (
	updateAllDir string
)

func init() {
	rootCmd.AddCommand(updateAllCmd)

	updateAllCmd.Flags().StringVarP(&updateAllDir, "directory", "d", ".", "Directory to search for encrypted files")
}

func runUpdateAll(cmd *cobra.Command, args []string) error {
	// Check if SOPS is installed
	if !sops.IsSopsInstalled() {
		return fmt.Errorf("SOPS is not installed. Please install SOPS first")
	}

	// Find all .sops files
	files, err := sops.FindSopsFiles(updateAllDir)
	if err != nil {
		return fmt.Errorf("failed to find encrypted files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No encrypted files found")
		return nil
	}

	fmt.Printf("Found %d encrypted files\n", len(files))
	fmt.Println("")

	var failed []string
	successCount := 0

	for i, file := range files {
		fmt.Printf("[%d/%d] Updating %s...\n", i+1, len(files), file)

		if err := sops.Updatekeys(file); err != nil {
			fmt.Printf("  ❌ Failed: %s\n", err)
			failed = append(failed, file)
		} else {
			fmt.Println("  ✓ Success")
			successCount++
		}
	}

	// Summary
	fmt.Println("")
	fmt.Printf("📋 Summary: %d/%d files updated successfully\n", successCount, len(files))

	if len(failed) > 0 {
		fmt.Printf("\n❌ Failed files:\n")
		for _, f := range failed {
			fmt.Printf("  - %s\n", f)
		}
		return fmt.Errorf("failed to update %d files", len(failed))
	}

	return nil
}
