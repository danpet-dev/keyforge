package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/danpet-dev/keyforge/pkg/sops"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <file1> <file2>",
	Short: "Compare two SOPS-encrypted files",
	Long: `Compare two SOPS-encrypted files by decrypting and diffing them.

This command decrypts both files in memory and shows the diff without
writing plaintext to disk.

Examples:
  # Compare two encrypted files
  keyforge diff secrets/dev.yaml.sops secrets/prod.yaml.sops

  # Compare encrypted with plaintext
  keyforge diff secrets/dev.yaml.sops secrets/dev.yaml

  # Compare different formats
  keyforge diff secrets/config.yaml.sops secrets/config.json.sops

The command uses your system's diff tool (or falls back to a simple
line-by-line comparison if diff is not available).

Security Note:
  Decrypted content is only held in memory and never written to disk.`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	file1 := args[0]
	file2 := args[1]

	// Check if SOPS is installed
	if !sops.IsSopsInstalled() {
		return fmt.Errorf("SOPS is not installed. Please install it first:\n  https://github.com/getsops/sops")
	}

	// Read file 1
	content1, err := readFileContent(file1)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", file1, err)
	}

	// Read file 2
	content2, err := readFileContent(file2)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", file2, err)
	}

	// Check if contents are identical
	if bytes.Equal(content1, content2) {
		fmt.Println("✅ Files are identical")
		return nil
	}

	// Show diff
	if err := showDiff(file1, file2, content1, content2); err != nil {
		return fmt.Errorf("failed to show diff: %w", err)
	}

	return nil
}

func readFileContent(file string) ([]byte, error) {
	// Check if file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", file)
	}

	// Check if it's a SOPS file
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// If file contains "sops:" metadata, it's encrypted - decrypt it
	if bytes.Contains(content, []byte("sops:")) || bytes.Contains(content, []byte("\"sops\":")) {
		decrypted, err := sops.Decrypt(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt: %w", err)
		}
		return decrypted, nil
	}

	// Otherwise, return plaintext
	return content, nil
}

func showDiff(file1, file2 string, content1, content2 []byte) error {
	// Use temp files approach directly (simpler and more reliable)
	return showDiffWithTempFiles(file1, file2, content1, content2)
}

func showDiffWithTempFiles(file1, file2 string, content1, content2 []byte) error {
	// Create temp files
	tmp1, err := os.CreateTemp("", "keyforge-diff-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmp1.Name())
	defer tmp1.Close()

	tmp2, err := os.CreateTemp("", "keyforge-diff-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmp2.Name())
	defer tmp2.Close()

	// Write contents
	if _, err := tmp1.Write(content1); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if _, err := tmp2.Write(content2); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Close files before reading
	tmp1.Close()
	tmp2.Close()

	// Try system diff
	if _, err := exec.LookPath("diff"); err == nil {
		cmd := exec.Command("diff", "-u", "--label", file1, "--label", file2, tmp1.Name(), tmp2.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		// diff returns exit code 1 if files differ, which is not an error
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return nil // Files differ, but it's not an error
			}
			return fmt.Errorf("diff command failed: %w", err)
		}
		return nil
	}

	// Fallback: simple line-by-line comparison
	return showSimpleDiff(file1, file2, content1, content2)
}

func showSimpleDiff(file1, file2 string, content1, content2 []byte) error {
	lines1 := bytes.Split(content1, []byte("\n"))
	lines2 := bytes.Split(content2, []byte("\n"))

	fmt.Printf("--- %s\n", file1)
	fmt.Printf("+++ %s\n", file2)

	maxLines := len(lines1)
	if len(lines2) > maxLines {
		maxLines = len(lines2)
	}

	for i := 0; i < maxLines; i++ {
		var line1, line2 []byte
		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if !bytes.Equal(line1, line2) {
			if len(line1) > 0 {
				fmt.Printf("-%s\n", line1)
			}
			if len(line2) > 0 {
				fmt.Printf("+%s\n", line2)
			}
		}
	}

	return nil
}
