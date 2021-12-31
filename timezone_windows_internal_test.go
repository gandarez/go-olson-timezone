//go:build windows

package timezone

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	err := os.Setenv("TZ", "America/Sao_Paulo")
	require.NoError(t, err)

	defer os.Unsetenv("TZ")

	tz := parseEnv()

	assert.Equal(t, "America/Sao_Paulo", tz)
}

func TestEnv_Filepath(t *testing.T) {
	tests := map[string]struct {
		Filepath        string
		Dir             string
		DestinationPath string
		Expected        string
	}{
		"America/Sao_Paulo": {
			Filepath:        "testdata/zoneinfo/America/Sao_Paulo",
			Dir:             "America",
			DestinationPath: "America/Sao_Paulo",
			Expected:        "America/Sao_Paulo",
		},
		"UTC": {
			Filepath:        "testdata/zoneinfo/UTC",
			Dir:             "UTC",
			DestinationPath: "UTC/UTC",
			Expected:        "UTC",
		},
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err = os.Mkdir(filepath.Join(tmpDir, test.Dir), os.FileMode(int(0700)))
			require.NoError(t, err)

			tmpTimezonePath := filepath.Join(tmpDir, test.DestinationPath)

			copyFile(t, test.Filepath, tmpTimezonePath)

			err = os.Setenv("TZ", tmpTimezonePath)
			require.NoError(t, err)

			defer os.Unsetenv("TZ")

			tz := parseEnv()

			assert.Equal(t, test.Expected, tz)
		})
	}
}

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}
