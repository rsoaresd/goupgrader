package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/rsoaresd/goupgrader/pkg/cmd/flags"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewGenerateConfigForOpenshiftDependencies() *cobra.Command {
	var targetOpenshiftVersion, currentOperatorSdkVersion, outputPath string

	command := &cobra.Command{
		Use:   "generate --target-openshift-version=<target-openshift-version> --in-use-op-sdk-version=<in-use-op-sdk-version> --output=<path-to-save-config>",
		Short: "Generate a YAML config with dependencies matching a target OpenShift version",
		Long: `This command analyzes the Kubernetes version used by a specific OpenShift release and compares it to the 
dependencies used by various operator-sdk versions. Once a compatible operator-sdk version is found (with matching 
Kubernetes minor version), it fetches related dependency versions and generates a YAML configuration file. 
This config can be used to align your Go project dependencies with the target OpenShift release.`,
		Args: cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			return GenerateConfigForOpenshiftDependencies(targetOpenshiftVersion, currentOperatorSdkVersion, outputPath)
		},
	}
	command.Flags().StringVarP(&targetOpenshiftVersion, "target-openshift-version", "t", "", "openshift version you wish to upgrade dependencies")
	flags.MustMarkRequired(command, "target-openshift-version")
	command.Flags().StringVarP(&currentOperatorSdkVersion, "in-use-op-sdk-version", "i", "", "current operator-sdk version in your Go project")
	flags.MustMarkRequired(command, "in-use-op-sdk-version")
	command.Flags().StringVarP(&outputPath, "output", "o", "", "path to  save the YAML config with the dependencies list for the target Openshift version")
	flags.MustMarkRequired(command, "output")

	return command
}

func GenerateConfigForOpenshiftDependencies(openshiftVersion, currentOperatorSdkVersion, configPath string) error {
	// find which k8s version Openshift is using
	k8sVersionUsedByOpenshift, err := getKubernetesVersionUsedByOpenshift(openshiftVersion)
	if err != nil {
		return err
	}
	log.Info().Msgf("k8s version used by Openshift %s: %s", openshiftVersion, k8sVersionUsedByOpenshift)

	// find which operator sdk version is using the target k8s version and build the config
	cfg, err := findMatchingOperatorSDKConfig(k8sVersionUsedByOpenshift, currentOperatorSdkVersion, openshiftVersion)
	if err != nil {
		return err
	}

	return saveConfigToFile(cfg, configPath)
}

func generateVersions(startMajorS, startMinorS string) ([]string, error) {
	minorCount := 4 // how many minors to include (e.g. 38, 39, 40, 41)
	patchCount := 5 // how many patch versions per minor
	var versions []string

	startMajor, err := strconv.Atoi(startMajorS)
	if err != nil {
		return nil, err
	}

	startMinor, err := strconv.Atoi(startMinorS)
	if err != nil {
		return nil, err
	}

	for i := 0; i < minorCount; i++ {
		currentMinor := startMinor + i
		for p := 0; p <= patchCount; p++ {
			version := fmt.Sprintf("v%d.%d.%d", startMajor, currentMinor, p)
			versions = append(versions, version)
		}
	}

	// reverse to be in desc order
	return reverseSlice(versions), nil
}

func reverseSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func hasSameMinorVersion(v1, v2 string) (bool, error) {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	if len(v1Parts) < 2 || len(v2Parts) < 2 {
		return false, fmt.Errorf("invalid version format")
	}

	major1, err1 := strconv.Atoi(v1Parts[0])
	minor1, err2 := strconv.Atoi(v1Parts[1])
	major2, err3 := strconv.Atoi(v2Parts[0])
	minor2, err4 := strconv.Atoi(v2Parts[1])

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return false, fmt.Errorf("invalid version numbers")
	}

	return major1 == major2 && minor1 == minor2, nil
}

func saveConfigToFile(cfg *Config, filename string) error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filename, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Info().Msgf("config saved to %s", filename)

	return nil
}

func getKubernetesVersionUsedByOpenshift(openshiftVersion string) (string, error) {
	return GetKubernetesVersion("openshift/api", fmt.Sprintf("release-%s", openshiftVersion), "k8s.io/api")
}

func generateCandidateSdkVersions(currentVersion string) ([]string, error) {
	currentVersion = strings.TrimPrefix(currentVersion, "v")
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid operator-sdk version format")
	}
	return generateVersions(parts[0], parts[1])
}

func buildDependencyConfig(repo, branch, openshiftVersion string) (Config, error) {
	config := Config{}

	dependencies := []string{
		"sigs.k8s.io/controller-runtime",
		"github.com/operator-framework/api",
		"github.com/operator-framework/operator-registry",
		"sigs.k8s.io/controller-tools",
	}

	for _, dep := range dependencies {
		version, err := GetKubernetesVersion(repo, branch, dep)
		if err != nil {
			return Config{}, fmt.Errorf("failed to get version for %s: %w", dep, err)
		}
		config.Dependencies = append(config.Dependencies, Dependency{
			Package: dep,
			Version: version,
		})
	}

	openshiftPackages := []string{
		"github.com/openshift/api",
		"github.com/openshift/library-go",
	}

	for _, opPkg := range openshiftPackages {
		config.Dependencies = append(config.Dependencies, Dependency{
			Package: opPkg,
			Branch:  fmt.Sprintf("release-%s", openshiftVersion),
		})
	}

	return config, nil
}

func findMatchingOperatorSDKConfig(k8sVersionUsedByOpenshift, currentOperatorSdkVersion, openshiftVersion string) (*Config, error) {
	sdkVersions, err := generateCandidateSdkVersions(currentOperatorSdkVersion)
	if err != nil {
		return nil, err
	}

	for _, version := range sdkVersions {
		k8sVersionUsedBySdk, err := GetKubernetesVersion("operator-framework/operator-sdk", version, "k8s.io/api")
		if err != nil {
			log.Info().Msgf("skipping SDK version %s: %v", version, err)
			continue
		}

		if same, _ := hasSameMinorVersion(k8sVersionUsedBySdk, k8sVersionUsedByOpenshift); same {
			log.Info().Msgf("match found! SDK %s uses Kubernetes %s", version, k8sVersionUsedBySdk)

			config, err := buildDependencyConfig("operator-framework/operator-sdk", version, openshiftVersion)
			if err != nil {
				return nil, fmt.Errorf("error building config: %w", err)
			}

			return &config, nil
		}
	}

	return nil, fmt.Errorf("no matching operator-sdk version found for Kubernetes %s", k8sVersionUsedByOpenshift)
}
