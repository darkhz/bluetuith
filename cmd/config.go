package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/darkhz/bluetuith/theme"
)

var (
	configPath       string
	configProperties map[string]string
)


// SetupConfig checks for the config directory,
// and creates one if it doesn't exist.
func SetupConfig() error {
	var dotConfigExists bool

	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configProperties = make(map[string]string)
	configDirs := []string{filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "bluetuith"), filepath.Join(homedir, ".bluetuith")}

	for i := range configDirs {
		if _, err := os.Stat(configDirs[i]); err == nil {
			configPath = configDirs[i]
			return theme.CreateThemesDir(configPath)
		}

		if i == 0 {
			if _, err := os.Stat(
				filepath.Clean(filepath.Dir(configDirs[i])),
			); err == nil {
				dotConfigExists = true
			}
		}
	}

	if configPath == "" {
		if dotConfigExists {
			err := os.Mkdir(configDirs[0], 0700)
			if err != nil {
				return fmt.Errorf("Cannot create %s", configDirs[0])
			}

			configPath = configDirs[0]
		} else {
			err := os.Mkdir(configDirs[1], 0700)
			if err != nil {
				return fmt.Errorf("Cannot create %s", configDirs[1])
			}

			configPath = configDirs[1]
		}
	}

	return theme.CreateThemesDir(configPath)
}

// ConfigPath returns the absolute path for the given configType.
func ConfigPath(configType string) (string, error) {
	confPath := filepath.Join(configPath, configType)

	if _, err := os.Stat(confPath); err != nil {
		fd, err := os.Create(confPath)
		fd.Close()
		if err != nil {
			return "", fmt.Errorf("Cannot create "+configType+" file at %s", confPath)
		}
	}

	return confPath, nil
}

// GetConfigProperty returns the value for the given property.
func GetConfigProperty(property string) string {
	return configProperties[property]
}

// AddConfigProperty adds a property and its value to the properties store.
func AddConfigProperty(property, value string) {
	configProperties[property] = value
}
