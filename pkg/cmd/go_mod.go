package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

var ErrPackageNotFound = errors.New("package not found")

type commandExecutor interface {
	Output() ([]byte, error)
	Run() error
}

var goCommandFunc = func(isStandard bool, projectPath string, arg ...string) commandExecutor {
	cmd := exec.Command("go", arg...)
	cmd.Dir = projectPath

	if isStandard {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd
}

func getPackageVersion(targetDir, packageName string) (string, error) {
	log.Info().Msgf("checking current version for package %s...", packageName)
	cmd := goCommandFunc(false, targetDir, "mod", "edit", "-json")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run 'go mod edit': %w", err)
	}

	var module Module
	if err := json.Unmarshal(output, &module); err != nil {
		return "", fmt.Errorf("failed to parse go.mod JSON: %w", err)
	}

	for _, pkg := range module.Require {
		if pkg.Path == packageName {
			return pkg.Version, nil
		}
	}

	return "", fmt.Errorf("%w: %s", ErrPackageNotFound, packageName)
}

// GetKubernetesVersion fetches the go.mod file from the given GitHub raw URL
// and returns the version of pkg used in that file.
func GetKubernetesVersion(repo, branch, pkg string) (string, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "raw.githubusercontent.com",
		Path:   fmt.Sprintf("%s/%s/go.mod", repo, branch),
	}
	resp, err := http.Get(u.String())

	if err != nil {
		return "", fmt.Errorf("failed to fetch go.mod: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non 200 response: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	inRequireBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// handle multiline require block
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}
		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		// check for pkg
		if strings.HasPrefix(line, pkg) {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1], nil
			}
			return "", fmt.Errorf("found %s but line is malformed: %q", pkg, line)
		}

		// also handle single-line require
		if strings.HasPrefix(line, "require") && strings.Contains(line, pkg) {
			// Example: require pkg v0.29.2
			fields := strings.Fields(line)
			for i, val := range fields {
				if val == pkg && i+1 < len(fields) {
					return fields[i+1], nil
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error: %w", err)
	}

	return "", fmt.Errorf("%s not found in go.mod", pkg)
}
