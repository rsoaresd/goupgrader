package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpgrade(t *testing.T) {
	tests := []struct {
		name              string
		config            string
		targetDir         string
		expectedError     string
		mockGoCommandFunc func(isStandard bool, projectPath string, arg ...string) commandExecutor
	}{
		{
			name: "upgrade working as expected",
			config: `dependencies:
  - package: "sigs.k8s.io/controller-runtime"
    version: "v0.19.3"
  - package: "github.com/openshift/api"
    branch: "release-4.18"`,
			targetDir:     "/path/to/project",
			expectedError: "",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: `{"Require":[{"Path":"sigs.k8s.io/controller-runtime","Version":"v0.19.3"},{"Path":"github.com/openshift/api","Version":"release-4.17"}]}`}
			},
		},
		{
			name: "invalid config",
			config: `dependencies:
  - package: "sigs.k8s.io/controller-runtime"
    version: "v0.19.3"
    branch: "release-4.18"`,
			targetDir:     "/path/to/project",
			expectedError: "dependency sigs.k8s.io/controller-runtime: cannot specify both version and branch",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{}
			},
		},
		{
			name: "package not found - should skip",
			config: `dependencies:
  - package: "sigs.k8s.io/controller-runtime"
    version: "v0.19.3"
  - package: "github.com/openshift/api"
    branch: "release-4.18"`,
			targetDir:     "/path/to/project",
			expectedError: "",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: `{"Require":[{"Path":"sigs.k8s.io/controller-runtime","Version":"v0.19.3"}]}`}
			},
		},
		{
			name: "error upgrading package - failed to get package version",
			config: `dependencies:
 - package: "sigs.k8s.io/controller-runtime"
   version: "v0.19.3"
 - package: "github.com/openshift/api"
   branch: "release-4.18"`,
			targetDir:     "/path/to/project",
			expectedError: "failed to parse go.mod JSON: unexpected end of JSON input",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: "", OutputErr: nil}
			},
		},
		{
			name: "error upgrading package - failed to run go get",
			config: `dependencies:
 - package: "sigs.k8s.io/controller-runtime"
   version: "v0.19.3"
 - package: "github.com/openshift/api"
   branch: "release-4.18"`,
			targetDir:     "/path/to/project",
			expectedError: "error upgrading dependency github.com/openshift/api: failed to run go get",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: `{"Require":[{"Path":"sigs.k8s.io/controller-runtime","Version":"v0.19.3"},{"Path":"github.com/openshift/api","Version":"release-4.17"}]}`, OutputErr: nil, RunErr: fmt.Errorf("failed to run go get")}
			},
		},
		{
			name: "no upgrade needed",
			config: `dependencies:
  - package: "sigs.k8s.io/controller-runtime"
    version: "v0.19.3"
  - package: "github.com/openshift/api"
    branch: "release-4.17"`,
			targetDir:     "/path/to/project",
			expectedError: "",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: `{"Require":[{"Path":"sigs.k8s.io/controller-runtime","Version":"v0.19.3"},{"Path":"github.com/openshift/api","Version":"release-4.17"}]}`}
			},
		},
		{
			name: "invalid branch",
			config: `dependencies:
  - package: "sigs.k8s.io/controller-runtime"
    version: "v0.19.3"
  - package: "github.com/openshift/api"
    branch: "release-4.017"`,
			targetDir:     "/path/to/project",
			expectedError: "no commit found for branch release-4.017",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: `{"Require":[{"Path":"sigs.k8s.io/controller-runtime","Version":"v0.19.3"},{"Path":"github.com/openshift/api","Version":"release-4.17"}]}`}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tmpFile, err := os.CreateTemp(tempDir, "config.yaml")
			require.NoError(t, err)

			_, err = tmpFile.WriteString(tt.config)
			require.NoError(t, err)

			goCommandFunc = tt.mockGoCommandFunc
			// err = Upgrade(tmpFile.Name(), tt.targetDir)

			cmd := NewUpgrade()
			args := []string{
				fmt.Sprintf("--config=%s", tmpFile.Name()),
				fmt.Sprintf("--project=%s", tt.targetDir),
			}
			cmd.SetArgs(args)

			err = cmd.Execute()

			if tt.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
