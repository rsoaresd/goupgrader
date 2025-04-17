package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

// getVersionWithCommitHashForBranch fetches the latest commit hash for the given branch in version format
func getVersionWithCommitHashForBranch(repo, branch string) (string, error) {
	repoURL := fmt.Sprintf("https://%s.git", repo)

	// get commit hash for the branch
	cmd := exec.Command("git", "ls-remote", repoURL, branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error fetching commit hash for branch %s: %w", branch, err)
	}
	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return "", fmt.Errorf("no commit found for branch %s", branch)
	}
	commitHash := fields[0]

	// strip github.com/ from the repo path
	apiRepoPath := strings.TrimPrefix(repo, "github.com/")

	u := &url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   fmt.Sprintf("repos/%s/commits/%s", apiRepoPath, commitHash),
	}
	resp, err := http.Get(u.String())

	if err != nil {
		return "", fmt.Errorf("failed to fetch commit info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API returned status: %s - %s", resp.Status, string(body))
	}

	var commit Commit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return "", fmt.Errorf("failed to decode GitHub API response: %w", err)
	}

	timestamp := commit.Commit.Committer.Date.UTC().Format("20060102150405")
	version := fmt.Sprintf("v0.0.0-%s-%s", timestamp, commitHash[:12])

	return version, nil
}
