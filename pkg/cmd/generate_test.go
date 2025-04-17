package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateConfigForOpenshiftDependencies(t *testing.T) {
	t.Run("generates config successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		tmpFile, err := os.CreateTemp(tempDir, "config")
		require.NoError(t, err)

		output := tmpFile.Name() + ".yaml"

		cmd := NewGenerateConfigForOpenshiftDependencies()
		args := []string{
			"--target-openshift-version=4.18",
			"--in-use-op-sdk-version=v1.39.0",
			fmt.Sprintf("--output=%s", output),
		}
		cmd.SetArgs(args)

		err = cmd.Execute()
		require.NoError(t, err)

		fileContent, err := os.ReadFile(output)
		require.NoError(t, err)

		expectedContent := `dependencies:
- package: sigs.k8s.io/controller-runtime
  version: v0.19.7
- package: github.com/operator-framework/api
  version: v0.27.0
- package: github.com/operator-framework/operator-registry
  version: v1.49.0
- package: sigs.k8s.io/controller-tools
  version: v0.16.5
- package: github.com/openshift/api
  branch: release-4.18
- package: github.com/openshift/library-go
  branch: release-4.18
`
		assert.Equal(t, expectedContent, string(fileContent))
	})

	testCases := []struct {
		name                  string
		args                  []string
		expectedError         string
		expectedOutputContent string
	}{
		{
			name: "wrong openshift version",
			args: []string{
				"--target-openshift-version=vvvv4.18",
				"--in-use-op-sdk-version=v1.39.0",
				"--output=/tmp/invalid.yaml",
			},
			expectedError: "non 200 response: 404",
		},
		{
			name: "invalid syntax in current SDK version",
			args: []string{
				"--target-openshift-version=4.18",
				"--in-use-op-sdk-version=vvvv1.39.0",
				"--output=/tmp/invalid.yaml",
			},
			expectedError: "strconv.Atoi: parsing \"vvv1\": invalid syntax",
		},
		{
			name: "invalid SDK version format",
			args: []string{
				"--target-openshift-version=4.18",
				"--in-use-op-sdk-version=v1.39.0.0",
				"--output=/tmp/invalid.yaml",
			},
			expectedError: "invalid operator-sdk version format",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			cmd := NewGenerateConfigForOpenshiftDependencies()
			cmd.SetArgs(test.args)

			err := cmd.Execute()

			if test.expectedError != "" {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
