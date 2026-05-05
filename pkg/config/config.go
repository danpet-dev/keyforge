package config

import (
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// SopsConfig represents the structure of .sops.yaml
type SopsConfig struct {
	CreationRules []CreationRule `yaml:"creation_rules"`
}

// CreationRule represents a single creation rule in .sops.yaml
type CreationRule struct {
	PathRegex      string      `yaml:"path_regex"`
	PGP            interface{} `yaml:"pgp,omitempty"`
	Age            interface{} `yaml:"age,omitempty"`
	EncryptedRegex string      `yaml:"encrypted_regex,omitempty"`
}

// GetPGPKeys returns PGP keys as a string slice (handles both string and []string)
func (r *CreationRule) GetPGPKeys() []string {
	return normalizeKeys(r.PGP)
}

// GetAgeKeys returns Age keys as a string slice (handles both string and []string)
func (r *CreationRule) GetAgeKeys() []string {
	return normalizeKeys(r.Age)
}

// normalizeKeys converts interface{} to []string (handles both string and []string)
func normalizeKeys(keys interface{}) []string {
	if keys == nil {
		return []string{}
	}

	switch v := keys.(type) {
	case string:
		// Handle comma-separated strings (common in YAML with folded scalars)
		if strings.Contains(v, ",") {
			parts := strings.Split(v, ",")
			result := []string{}
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					result = append(result, trimmed)
				}
			}
			return result
		}
		return []string{strings.TrimSpace(v)}
	case []interface{}:
		result := make([]string, len(v))
		for i, k := range v {
			if str, ok := k.(string); ok {
				result[i] = strings.TrimSpace(str)
			}
		}
		return result
	case []string:
		return v
	default:
		return []string{}
	}
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
				existing := c.CreationRules[i].GetPGPKeys()
				existing = append(existing, key)
				c.CreationRules[i].PGP = existing
			case "age":
				existing := c.CreationRules[i].GetAgeKeys()
				existing = append(existing, key)
				c.CreationRules[i].Age = existing
			default:
				return fmt.Errorf("unknown key type: %s", keyType)
			}
			return nil
		}
	}

	return fmt.Errorf("path regex not found: %s", pathRegex)
}

// RemoveKey removes a PGP or Age key from all creation rules
// Returns a map of path_regex -> bool indicating which rules were affected
func (c *SopsConfig) RemoveKey(keyType, key string) map[string]bool {
	affectedRules := make(map[string]bool)

	for i := range c.CreationRules {
		removed := false

		switch keyType {
		case "pgp":
			existing := c.CreationRules[i].GetPGPKeys()
			filtered := []string{}
			for _, k := range existing {
				if k != key {
					filtered = append(filtered, k)
				} else {
					removed = true
				}
			}
			if removed {
				if len(filtered) > 0 {
					c.CreationRules[i].PGP = filtered
				} else {
					c.CreationRules[i].PGP = nil
				}
				affectedRules[c.CreationRules[i].PathRegex] = true
			}

		case "age":
			existing := c.CreationRules[i].GetAgeKeys()
			filtered := []string{}
			for _, k := range existing {
				if k != key {
					filtered = append(filtered, k)
				} else {
					removed = true
				}
			}
			if removed {
				if len(filtered) > 0 {
					c.CreationRules[i].Age = filtered
				} else {
					c.CreationRules[i].Age = nil
				}
				affectedRules[c.CreationRules[i].PathRegex] = true
			}
		}
	}

	return affectedRules
}

// FindKeyByEmail finds a PGP key in the config by email
func (c *SopsConfig) FindKeyByEmail(email string) (string, bool) {
	// This requires matching against keys.ListPGPKeys()
	// We'll implement this in the command layer
	return "", false
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
