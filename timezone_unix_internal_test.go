//go:build darwin || linux

package timezone

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected string
	}{
		"inline": {
			Filepath: "testdata/var/db/zoneinfo",
			Expected: "America/Sao_Paulo",
		},
		"with comment": {
			Filepath: "testdata/etc/timezone_comment",
			Expected: "America/Sao_Paulo",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tz := parseFromConfigFile([]string{test.Filepath})

			assert.Equal(t, []string{test.Expected}, tz)
		})
	}
}

func TestParse_Empty(t *testing.T) {
	tz := parseFromConfigFile([]string{"testdata/empty"})

	assert.Equal(t, []string{}, tz)
}

func TestParseFromClock(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected string
	}{
		"timezone": {
			Filepath: "testdata/etc/conf.d/clock",
			Expected: "America/Sao_Paulo",
		},
		"zone": {
			Filepath: "testdata/etc/sysconfig/clock",
			Expected: "America/Sao_Paulo",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tz := parseFromClock([]string{test.Filepath})

			assert.Equal(t, []string{test.Expected}, tz)
		})
	}
}

func TestParseFromClock_Empty(t *testing.T) {
	tz := parseFromClock([]string{"testdata/empty"})

	assert.Equal(t, []string{}, tz)
}

func TestParseSymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	symlinkFile := filepath.Join(tmpDir, "localtime")

	err = os.Symlink("testdata/zoneinfo/America/Sao_Paulo", symlinkFile)
	require.NoError(t, err)

	tz := parseSymlink(symlinkFile)

	assert.Equal(t, "America/Sao_Paulo", tz)
}

func TestParseSymlink_NotSymlink(t *testing.T) {
	tz := parseSymlink("testdata/zoneinfo/America/Sao_Paulo")

	assert.Empty(t, tz)
}

func TestResolveTimezones(t *testing.T) {
	tests := map[string]struct {
		Timezones []string
		Expected  string
	}{
		"no timezone": {
			Timezones: []string{},
			Expected:  "",
		},
		"only one timezone": {
			Timezones: []string{"America/Sao_Paulo"},
			Expected:  "America/Sao_Paulo",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tz, err := resolveTimezones(test.Timezones, "")
			require.NoError(t, err)

			assert.Equal(t, test.Expected, tz)
		})
	}
}

func TestResolveTimezones_Conflicting(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	timezones := []string{
		"America/Sao_Paulo",
		"America/Los_Angeles",
		"Africa/Harare",
	}

	_, err = resolveTimezones(timezones, "testdata/zoneinfo")
	require.Error(t, err)

	assert.Contains(
		t,
		err.Error(),
		"multiple conflicting time zone configurations found:\n",
		fmt.Sprintf("error %q differs from the string set", err))

	assert.Contains(
		t,
		err.Error(),
		"America/Sao_Paulo\n",
		fmt.Sprintf("error %q differs from the string set", err))

	assert.Contains(
		t,
		err.Error(),
		"America/Los_Angeles\n",
		fmt.Sprintf("error %q differs from the string set", err))

	assert.Contains(
		t,
		err.Error(),
		"Africa/Harare\n",
		fmt.Sprintf("error %q differs from the string set", err))

	assert.Contains(
		t,
		err.Error(),
		"Fix the configuration, or set the time zone in a TZ environment variable",
		fmt.Sprintf("error %q differs from the string set", err))
}

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
