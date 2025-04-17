package cmd

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// parseConfig parses the YAML configuration file at the given path
// and validates the dependencies specified in the file.
func parseConfig(configPath string) (*Config, error) {
	// open and read the YAML file
	yamlFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer yamlFile.Close()

	// parse the YAML file
	var config Config
	decoder := yaml.NewDecoder(yamlFile)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	// validate config
	for _, dep := range config.Dependencies {
		if err := validateDependency(dep); err != nil {
			return nil, err
		}
	}

	return &config, nil
}

// validateDependency checks if a dependency has valid version or branch attributes.
func validateDependency(dependency Dependency) error {
	// check if both version and branch are provided, which is invalid
	if dependency.Version != "" && dependency.Branch != "" {
		return fmt.Errorf("dependency %s: cannot specify both version and branch", dependency.Package)
	}

	// ensure at least one of version or branch is provided
	if dependency.Version == "" && dependency.Branch == "" {
		return fmt.Errorf("dependency %s: must specify either version or branch", dependency.Package)
	}

	// if version is specified, it should be a non-empty string
	if dependency.Version != "" && strings.TrimSpace(dependency.Version) == "" {
		return fmt.Errorf("dependency %s: version cannot be an empty string", dependency.Package)
	}

	// if branch is specified, it should be a non-empty string
	if dependency.Branch != "" && strings.TrimSpace(dependency.Branch) == "" {
		return fmt.Errorf("dependency %s: branch cannot be an empty string", dependency.Package)
	}

	return nil
}
