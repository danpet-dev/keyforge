package commands

import (
	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "keyforge",
	Short: "SOPS multi-key lifecycle management",
	Long: `KeyForge - Forge your encryption keys with confidence

KeyForge simplifies SOPS multi-key management by automating common tasks
like key rotation, team onboarding, and .sops.yaml configuration.

Perfect for teams managing encrypted secrets across multiple environments.`,
	Version: Version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
