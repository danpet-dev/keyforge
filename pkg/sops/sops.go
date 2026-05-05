package sops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Updatekeys runs sops updatekeys on a file with proper format flags
func Updatekeys(file string) error {
	if !strings.HasSuffix(file, ".sops") {
		return fmt.Errorf("file must end with .sops: %s", file)
	}

	// Auto-detect input type
	inputType := "yaml"
	if strings.Contains(file, ".json.sops") {
		inputType = "json"
	} else if strings.Contains(file, ".env.sops") {
		inputType = "dotenv"
	}

	cmd := exec.Command("sops", "updatekeys", "--input-type", inputType, "--yes", file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sops updatekeys failed for %s: %w", file, err)
	}

	return nil
}

// Edit opens a SOPS file in the editor
func Edit(file string) error {
	inputType := "yaml"
	if strings.Contains(file, ".json.sops") {
		inputType = "json"
	} else if strings.Contains(file, ".env.sops") {
		inputType = "dotenv"
	}

	cmd := exec.Command("sops", "--input-type", inputType, "--output-type", inputType, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sops edit failed for %s: %w", file, err)
	}

	return nil
}

// FindSopsFiles finds all *.sops files in a directory recursively
func FindSopsFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".sops") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", root, err)
	}

	return files, nil
}

// IsSopsInstalled checks if sops is available in PATH
func IsSopsInstalled() bool {
	_, err := exec.LookPath("sops")
	return err == nil
}

// GetSopsVersion returns the installed SOPS version
func GetSopsVersion() (string, error) {
	cmd := exec.Command("sops", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get sops version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// Decrypt decrypts a SOPS file and returns the plaintext content
func Decrypt(file string) ([]byte, error) {
	// Auto-detect input type
	inputType := "yaml"
	if strings.Contains(file, ".json.sops") {
		inputType = "json"
	} else if strings.Contains(file, ".env.sops") {
		inputType = "dotenv"
	}

	cmd := exec.Command("sops", "--decrypt", "--input-type", inputType, "--output-type", inputType, file)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("sops decrypt failed for %s: %w\n%s", file, err, string(output))
	}

	return output, nil
}

// Encrypt encrypts a plaintext file with SOPS
func Encrypt(file string) error {
	// Auto-detect input type
	inputType := "yaml"
	if strings.HasSuffix(file, ".json") {
		inputType = "json"
	} else if strings.HasSuffix(file, ".env") {
		inputType = "dotenv"
	}

	outputFile := file + ".sops"

	cmd := exec.Command("sops", "--encrypt", "--input-type", inputType, "--output", outputFile, file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sops encrypt failed for %s: %w", file, err)
	}

	return nil
}
