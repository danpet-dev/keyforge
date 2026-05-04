package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SopsConfig represents the structure of .sops.yaml
type SopsConfig struct {
	CreationRules []CreationRule `yaml:"creation_rules"`
}

// CreationRule represents a single creation rule in .sops.yaml
type CreationRule struct {
	PathRegex      string   `yaml:"path_regex"`
	PGP            []string `yaml:"pgp,omitempty"`
	Age            []string `yaml:"age,omitempty"`
	EncryptedRegex string   `yaml:"encrypted_regex,omitempty"`
}

// Load reads and parses .sops.yaml from the given path
func Load(path string) (*SopsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var config SopsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return &config, nil
}

// Save writes the config to the specified path
func (c *SopsConfig) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

// AddKey adds a PGP or Age key to a specific path regex
func (c *SopsConfig) AddKey(pathRegex, keyType, key string) error {
	for i := range c.CreationRules {
		if c.CreationRules[i].PathRegex == pathRegex {
			switch keyType {
			case "pgp":
				c.CreationRules[i].PGP = append(c.CreationRules[i].PGP, key)
			case "age":
				c.CreationRules[i].Age = append(c.CreationRules[i].Age, key)
			default:
				return fmt.Errorf("unknown key type: %s", keyType)
			}
			return nil
		}
	}

	return fmt.Errorf("path regex not found: %s", pathRegex)
}

// GenerateTemplate creates a new .sops.yaml with best-practice structure
func GenerateTemplate(projectName string, keyType string, keys map[string][]string) *SopsConfig {
	config := &SopsConfig{
		CreationRules: []CreationRule{},
	}

	// Environment-specific rules
	environments := []string{"development", "test", "prod"}
	for _, env := range environments {
		rule := CreationRule{
			PathRegex:      fmt.Sprintf("secrets/%s\\.yaml\\.sops$", env),
			EncryptedRegex: "^(password|secret|key|token|credentials|private)$",
		}

		if keyType == "pgp" {
			rule.PGP = keys[env]
		} else if keyType == "age" {
			rule.Age = keys[env]
		}

		config.CreationRules = append(config.CreationRules, rule)
	}

	// Service-specific secrets
	rule := CreationRule{
		PathRegex:      `services/.*/k8s/secrets/.*\.yaml\.sops$`,
		EncryptedRegex: "^(password|secret|key|token|credentials|private)$",
	}
	if keyType == "pgp" {
		rule.PGP = keys["all"]
	} else if keyType == "age" {
		rule.Age = keys["all"]
	}
	config.CreationRules = append(config.CreationRules, rule)

	// Database secrets
	dbRule := CreationRule{
		PathRegex:      `database/.*/secrets/.*\.yaml\.sops$`,
		EncryptedRegex: "^(password|secret|key|token|credentials|private)$",
	}
	if keyType == "pgp" {
		dbRule.PGP = keys["all"]
	} else if keyType == "age" {
		dbRule.Age = keys["all"]
	}
	config.CreationRules = append(config.CreationRules, dbRule)

	return config
}
