package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersionWithCommitHashForBranch(t *testing.T) {
	tests := []struct {
		name            string
		repository      string
		branch          string
		expectedVersion string
		expectedError   string
	}{
		{
			name:            "valid repository and branch",
			repository:      "github.com/openshift/api",
			branch:          "release-4.18",
			expectedVersion: "v0.0.0-20250410062700-d6c84c55a124",
			expectedError:   "",
		},
		{
			name:            "invalid repository",
			repository:      "github.com/openshift000/api",
			branch:          "release-4.18",
			expectedVersion: "",
			expectedError:   "error fetching commit hash for branch release-4.18: exit status 128",
		},
		{
			name:            "invalid branch",
			repository:      "github.com/openshift/api",
			branch:          "release-4.018",
			expectedVersion: "",
			expectedError:   "no commit found for branch release-4.018",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version, err := getVersionWithCommitHashForBranch(test.repository, test.branch)

			if test.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, test.expectedVersion, version)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}
