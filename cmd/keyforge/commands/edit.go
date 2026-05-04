package commands

import (
	"fmt"

	"github.com/danpet-dev/keyforge/pkg/sops"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <file>",
	Short: "Edit encrypted file with SOPS",
	Long: `Edit an encrypted file with proper format detection.

This is a convenience wrapper around 'sops' that automatically detects
the file format (yaml, json, env) and passes the correct flags.

Example:
  keyforge edit secrets/development.yaml.sops
  keyforge edit config.json.sops
  keyforge edit .env.sops`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	file := args[0]

	// Check if SOPS is installed
	if !sops.IsSopsInstalled() {
		return fmt.Errorf("SOPS is not installed. Please install SOPS first")
	}

	// Edit file
	if err := sops.Edit(file); err != nil {
		return err
	}

	fmt.Printf("✓ Saved %s\n", file)
	return nil
}
