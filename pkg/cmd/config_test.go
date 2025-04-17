package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	t.Run("valid config file", func(t *testing.T) {
		config := `dependencies:
  - package: "sigs.k8s.io/controller-runtime"
    version: "v0.19.3"
  - package: "github.com/openshift/api"
    branch: "release-4.18"`

		// write the config to a temporary file
		tmpFile, err := os.CreateTemp("", "config.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name()) // clean up

		_, err = tmpFile.WriteString(config)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close()) // make sure it's flushed and closed

		parsedConfig, err := parseConfig(tmpFile.Name())
		require.NoError(t, err)

		assert.Len(t, parsedConfig.Dependencies, 2)
		assert.Equal(t, "sigs.k8s.io/controller-runtime", parsedConfig.Dependencies[0].Package)
		assert.Equal(t, "v0.19.3", parsedConfig.Dependencies[0].Version)
		assert.Empty(t, parsedConfig.Dependencies[0].Branch)
		assert.Equal(t, "github.com/openshift/api", parsedConfig.Dependencies[1].Package)
		assert.Empty(t, parsedConfig.Dependencies[1].Version)
		assert.Equal(t, "release-4.18", parsedConfig.Dependencies[1].Branch)
	})

	t.Run("invalid config file", func(t *testing.T) {
		config := ``

		// write the config to a temporary file
		tmpFile, err := os.CreateTemp("", "config.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name()) // clean up

		_, err = tmpFile.WriteString(config)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close()) // make sure it's flushed and closed

		_, err = parseConfig(tmpFile.Name())
		require.Error(t, err)
	})

	t.Run("non-existent config path", func(t *testing.T) {
		_, err := parseConfig("/nonexistent/path.yaml")
		require.Error(t, err)
	})
}

func TestValidateDependency(t *testing.T) {
	tests := []struct {
		name       string
		dependency Dependency
		expected   string
	}{
		// valid cases
		{
			name:       "Valid version only",
			dependency: Dependency{"package1", "v1.0.0", ""},
			expected:   "",
		},
		{
			name:       "Valid branch only",
			dependency: Dependency{"package2", "", "branch1"},
			expected:   "",
		},

		// invalid cases
		{
			name:       "Invalid: both version and branch set",
			dependency: Dependency{"package3", "v1.0.0", "branch1"},
			expected:   "dependency package3: cannot specify both version and branch",
		},
		{
			name:       "Invalid: neither version nor branch",
			dependency: Dependency{"package4", "", ""},
			expected:   "dependency package4: must specify either version or branch",
		},
		{
			name:       "Invalid: empty branch",
			dependency: Dependency{"package5", "", " "},
			expected:   "dependency package5: branch cannot be an empty string",
		},
		{
			name:       "Invalid: empty version",
			dependency: Dependency{"package6", " ", ""},
			expected:   "dependency package6: version cannot be an empty string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDependency(test.dependency)

			if test.expected == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expected)
			}
		})
	}
}
