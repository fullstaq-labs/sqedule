package lib

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/mitchellh/go-homedir"
)

// UserConfigDir returns the default root directory to use for user-specific configuration data.
// Users should create their own application-specific subdirectory within this one and use that.
//
// On Unix systems, it returns $XDG_CONFIG_HOME as specified by
// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html if
// non-empty, else $HOME/.config.
// On macOS, it returns $HOME/Library/Application Support.
// On Windows, it returns %AppData%.
// On Plan 9, it returns $home/lib.
//
// If the location cannot be determined (for example, $HOME is not defined),
// then it will return an error.
//
// The differences between this and `os.UserConfigDir` are:
//
// - on macOS we return $XDG_CONFIG_HOME or ~/.config instead of ~/Library/Application Support.
//
// - We determine the home directory through more than just $HOME, so we're resilient against cases in which $HOME is not set.
func UserConfigDir() (string, error) {
	switch runtime.GOOS {
	case "windows", "ios", "plan9":
		return os.UserConfigDir()

	default:
		result := os.Getenv("XDG_CONFIG_HOME")
		if len(result) > 0 {
			return result, nil
		}

		home, err := homedir.Dir()
		if err != nil {
			return "", fmt.Errorf("Error querying location of home directory: %w", err)
		}

		return path.Join(home, ".config"), nil
	}
}
