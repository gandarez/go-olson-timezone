//go:build windows

package timezone

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Name tries to find the local timezone configuration. Windows is special.
// It has unique time zone names (in several meanings of the word) available,
// but unfortunately, they can be translated to the language of the operating system,
// so we need to do a backwards lookup, by going through all time zones and see which
// one matches.
func Name() (string, error) {
	// first try the ENV setting
	if tzenv := parseEnv(); tzenv != "" {
		return tzenv, nil
	}

	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\TimeZoneInformation`,
		registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("failed to open registry key")
	}

	defer key.Close()

	tzwin, _, err := key.GetStringValue("TimeZoneKeyName")
	if err != nil {
		return "", fmt.Errorf("can not find windows timezone configuration")
	}

	// for some reason this returns a string with loads of NULL bytes at
	// least on some systems. I don't know if this is a bug somewhere, I
	// just work around it
	tzwin = strings.ReplaceAll(tzwin, "\x00", "")

	var tz string

	tz, ok := windowsTimezones[tzwin]
	if !ok {
		// try adding "Standard Time", it seems to work a lot of times
		tzwin += " Standard Time"
		tz, ok = windowsTimezones[tzwin]
	}

	if !ok {
		return "", fmt.Errorf("windows timezone '%s' not found", tzwin)
	}

	return tz, nil
}

// parseEnv parses timezone from TZ env var.
func parseEnv() string {
	tzenv := os.Getenv("TZ")
	if tzenv == "" {
		return ""
	}

	if _, ok := timezones[tzenv]; ok {
		return tzenv
	}

	if filepath.IsAbs(tzenv) && fileExists(tzenv) {
		// it's a file specification
		parts := strings.Split(tzenv, string(os.PathSeparator))

		// is it a zone info zone?
		joined := strings.Join(parts[len(parts)-2:], "/")
		if _, ok := timezones[joined]; ok {
			return joined
		}

		// maybe it's a short one, like UTC?
		if _, ok := timezones[parts[len(parts)-1]]; ok {
			return parts[len(parts)-1]
		}
	}

	return ""
}

// fileExists checks if a file or directory exist.
func fileExists(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}
