//go:build darwin || linux

package timezone

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yookoala/realpath"
)

var timezoneRegex = regexp.MustCompile(`^\s*(TIMEZONE|ZONE)\s*=\s*\"(?P<tz>.*)\"$`)

// Name tries to find the local timezone configuration. It returns the timezone name
// if found. If not, an error is returned.
func Name() (string, error) {
	// first try the ENV setting
	if tzenv := parseEnv(); tzenv != "" {
		return tzenv, nil
	}

	// now look for distribution specific configuration files
	// that contain the timezone name.
	timezones := []string{}

	timezones = append(timezones, parseFromConfigFile([]string{
		"/etc/timezone",
		"/var/db/zoneinfo"})...)

	timezones = append(timezones, parseFromClock([]string{
		"/etc/sysconfig/clock",
		"/etc/conf.d/clock"})...)

	parsed := parseSymlink("/etc/localtime")
	if parsed != "" {
		timezones = append(timezones, parsed)
	}

	return resolveTimezones(timezones, "/usr/share/zoneinfo")
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

// parse parses timezone from configuration files.
func parseFromConfigFile(paths []string) []string {
	timezones := []string{}

	for _, configfile := range paths {
		data, err := os.ReadFile(configfile)
		if err != nil {
			continue
		}

		etctz := strings.Trim(string(data), "/ \t\r\n")
		if etctz == "" {
			continue
		}

		lines := strings.Split(strings.ReplaceAll(etctz, "\r\n", "\n"), "\n")
		for _, line := range lines {
			// get rid of host definitions and comments
			if strings.Contains(line, " ") {
				line = strings.SplitN(line, " ", 2)[0]
			}

			if strings.Contains(line, "#") {
				line = strings.SplitN(line, "#", 2)[0]
			}

			line = strings.TrimSpace(line)

			if line != "" {
				timezones = append(timezones, strings.ReplaceAll(line, " ", "_"))
			}
		}
	}

	return timezones
}

// parseFromClock parses timezone from clock files.
// CentOS has a ZONE setting in /etc/sysconfig/clock,
// OpenSUSE has a TIMEZONE setting in /etc/sysconfig/clock and
// Gentoo has a TIMEZONE setting in /etc/conf.d/clock.
func parseFromClock(paths []string) []string {
	timezones := []string{}

	for _, filename := range paths {
		file, err := os.Open(filename)
		if err != nil {
			continue
		}

		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			// look for the TIMEZONE|ZONE= setting
			match := timezoneRegex.FindStringSubmatch(line)
			paramsMap := make(map[string]string)

			for i, name := range timezoneRegex.SubexpNames() {
				if i > 0 && i <= len(match) {
					paramsMap[name] = match[i]
				}
			}

			if len(paramsMap) == 0 || paramsMap["tz"] == "" {
				continue
			}

			timezones = append(timezones, strings.ReplaceAll(paramsMap["tz"], " ", "_"))
		}
	}

	return timezones
}

// parseSymlink parses symbolic link to resolve timezone name.
// The systemd distributions use symlinks that include the zone name.
func parseSymlink(path string) string {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return ""
	}

	// is symlink?
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		tz, err := os.Readlink(path)
		if err != nil {
			return ""
		}

		idx := strings.Index(tz, "zoneinfo/")
		if idx == -1 {
			return ""
		}

		return strings.ReplaceAll(tz[idx+9:], " ", "_")
	}

	return ""
}

// resolveTimezones resolves conflicted timezones. Otherwise returns an error.
func resolveTimezones(timezones []string, zoneinfo string) (string, error) {
	if len(timezones) == 0 {
		return "", nil
	}

	if len(timezones) == 1 {
		return timezones[0], nil
	}

	// multiple configs. See if they match
	var filtered []string

	depth := len(strings.Split(zoneinfo, string(os.PathSeparator)))

	for _, tzname := range timezones {
		// look them up in '/usr/share/zoneinfo', and find what they really point to
		path, err := realpath.Realpath(filepath.Join(zoneinfo, tzname))
		if err != nil {
			continue
		}

		name := strings.Join(strings.Split(path, string(os.PathSeparator))[depth:], "/")
		filtered = appendIfMissing(filtered, name)
	}

	if len(filtered) == 1 {
		return filtered[0], nil
	}

	if len(filtered) > 1 {
		message := "multiple conflicting time zone configurations found:\n"
		for _, v := range filtered {
			message += fmt.Sprintf("%s\n", v)
		}

		message += "Fix the configuration, or set the time zone in a TZ environment variable"

		return "", errors.New(message)
	}

	return "", nil
}

func appendIfMissing(slice []string, s string) []string {
	for _, item := range slice {
		if item == s {
			return slice
		}
	}

	return append(slice, s)
}

// fileExists checks if a file or directory exist.
func fileExists(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}
