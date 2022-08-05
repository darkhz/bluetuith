package cmd

import (
	"fmt"
	"os"
	"path/filepath"
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
	configDirs := []string{".config/bluetuith", ".bluetuith"}

	for i, dir := range configDirs {
		fullpath := filepath.Join(homedir, dir)
		configDirs[i] = fullpath

		if _, err := os.Stat(fullpath); err == nil {
			configPath = fullpath
			return nil
		}

		if i == 0 {
			if _, err := os.Stat(
				filepath.Clean(filepath.Dir(fullpath)),
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
		} else {
			err := os.Mkdir(configDirs[1], 0700)
			if err != nil {
				return fmt.Errorf("Cannot create %s", configDirs[1])
			}
		}
	}

	return nil
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
