package cmd

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/rsoaresd/goupgrader/pkg/cmd/flags"
	"github.com/spf13/cobra"
)

func NewUpgrade() *cobra.Command {
	var config, project string

	command := &cobra.Command{
		Use:   "upgrade --config=<config-path> --project=<project-path>",
		Short: "Upgrades your Go project dependencies based on a config file",
		Long: `Upgrades Go project dependencies based on the provided YAML config file.
Each dependency can define a version or a branch, and the tool will apply the appropriate upgrade.`,
		Args: cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			return Upgrade(config, project)
		},
	}

	command.Flags().StringVarP(&config, "config", "c", "", "path to YAML config")
	flags.MustMarkRequired(command, "config")
	command.Flags().StringVarP(&project, "project", "p", "", "path to the target Go project")
	flags.MustMarkRequired(command, "project")

	return command
}

// Upgrade performs the upgrade of project dependencies based on the provided configuration.
// It takes in two parameters:
// - configPath: The file path to the YAML configuration that contains dependency details.
// - projectPath: The file path to the Go project that needs the upgrades.
//
// The function does the following:
// 1. It parses the configuration file using `parseConfig`, which returns a list of dependencies to upgrade.
// 2. It iterates over each dependency in the configuration:
//   - If the dependency has a specified version, it calls `upgradePackage` to upgrade that package to the given version.
//   - If the dependency specifies a branch, it fetches the corresponding version (commit hash) for that branch using `getVersionWithCommitHashForBranch`, and then upgrades the package to that version.
//
// 3. If any errors are encountered during the upgrade process (either parsing the config, upgrading a package, or fetching a branch version), it returns the error.
// 4. Once all dependencies have been processed successfully, it returns `nil`, indicating the upgrade process is complete.
func Upgrade(configPath, projectPath string) error {
	config, err := parseConfig(configPath)
	if err != nil {
		return err
	}

	for _, dependency := range config.Dependencies {
		if dependency.Version != "" {
			err := upgradePackage(projectPath, dependency.Package, dependency.Version)
			if err != nil {
				return err
			}
		}

		if dependency.Branch != "" {
			targetVersion, err := getVersionWithCommitHashForBranch(dependency.Package, dependency.Branch)
			if err != nil {
				return err
			}

			err = upgradePackage(projectPath, dependency.Package, targetVersion)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func upgradePackage(projectPath, packageName, targetVersion string) error {
	currentVersion, err := getPackageVersion(projectPath, packageName)
	if err != nil {
		if errors.Is(err, ErrPackageNotFound) {
			log.Info().Msgf("skipping %s: not found in go.mod", packageName)
			return nil
		}
		return err
	}

	log.Info().Msgf("upgrading %s from %s to %s...", packageName, currentVersion, targetVersion)

	// if the current version is lower than the target version, upgrade
	if currentVersion < targetVersion {
		// upgrade package
		cmd := goCommandFunc(true, projectPath, "get", fmt.Sprintf("%s@%s", packageName, targetVersion))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error upgrading dependency %s: %w", packageName, err)
		}

		// run go mod tidy
		cmd = goCommandFunc(true, projectPath, "mod", "tidy")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running go mod tidy: %w", err)
		}

		log.Info().Msgf("upgrade %s from %s to %s finished successfully", packageName, currentVersion, targetVersion)

	} else {
		log.Info().Msgf("no upgrade needed for %s: current version %s >= requested version %s",
			packageName, currentVersion, targetVersion)
	}

	return nil
}
