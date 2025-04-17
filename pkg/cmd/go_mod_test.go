package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (m *MockCommandExecutor) Output() ([]byte, error) {
	return []byte(m.Outcome), m.OutputErr
}

func (m *MockCommandExecutor) Run() error {
	return m.RunErr
}

func TestGetPackageVersion(t *testing.T) {
	tests := []struct {
		name              string
		targetDir         string
		packageName       string
		expectedResult    string
		expectedError     string
		mockGoCommandFunc func(isStandard bool, projectPath string, arg ...string) commandExecutor
	}{
		{
			name:           "Package found in go.mod",
			targetDir:      "/path/to/project",
			packageName:    "sigs.k8s.io/controller-runtime",
			expectedResult: "v0.19.3",
			expectedError:  "",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: `{"Require":[{"Path":"sigs.k8s.io/controller-runtime","Version":"v0.19.3"}]}`}
			},
		},
		{
			name:           "Package not found in go.mod",
			targetDir:      "/path/to/project",
			packageName:    "github.com/openshift/api",
			expectedResult: "",
			expectedError:  "package not found: github.com/openshift/api",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: `{"Require":[{"Path":"sigs.k8s.io/controller-runtime","Version":"v0.19.3"}]}`}
			},
		},
		{
			name:           "Error running command",
			targetDir:      "/path/to/project",
			packageName:    "github.com/example/package",
			expectedResult: "",
			expectedError:  "failed to run 'go mod edit': failed to run 'go mod edit'",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: "", OutputErr: fmt.Errorf("failed to run 'go mod edit'")}
			},
		},
		{
			name:           "Error parsing",
			targetDir:      "/path/to/project",
			packageName:    "github.com/example/package",
			expectedResult: "",
			expectedError:  "failed to parse go.mod JSON: unexpected end of JSON input",
			mockGoCommandFunc: func(_ bool, _ string, _ ...string) commandExecutor {
				return &MockCommandExecutor{Outcome: "", OutputErr: nil}
			},
		},
	}

	origShellCommandFunc := goCommandFunc
	defer func() { goCommandFunc = origShellCommandFunc }()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			goCommandFunc = test.mockGoCommandFunc
			result, err := getPackageVersion(test.targetDir, test.packageName)

			if test.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, test.expectedResult, result)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}
