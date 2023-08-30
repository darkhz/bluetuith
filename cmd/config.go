package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hjson/hjson-go/v4"
	"github.com/knadh/koanf/v2"
)

// Config describes the configuration for the app.
type Config struct {
	path string

	*koanf.Koanf
}

// ConfigDir describes the path and properties of a configuration directory.
type ConfigDir struct {
	path, fullpath string

	exist, hidden, prefixHomeDir bool
}

var config Config

// setup checks for the config directory,
// and creates one if it doesn't exist.
func (c *Config) setup() {
	c.Koanf = koanf.New(".")

	homedir, err := os.UserHomeDir()
	if err != nil {
		PrintError(err.Error())
	}

	configPaths := []*ConfigDir{
		{path: os.Getenv("XDG_CONFIG_HOME")},
		{path: ".config", prefixHomeDir: true},
		{path: ".", hidden: true, prefixHomeDir: true},
	}

	for _, dir := range configPaths {
		name := "bluetuith"

		if dir.path == "" {
			continue
		}

		if dir.hidden {
			name = "." + name
		}

		if dir.prefixHomeDir {
			dir.path = filepath.Join(homedir, dir.path)
		}

		if _, err := os.Stat(filepath.Clean(dir.path)); err == nil {
			dir.exist = true
		}

		dir.fullpath = filepath.Join(dir.path, name)
		if _, err := os.Stat(filepath.Clean(dir.fullpath)); err == nil {
			c.path = dir.fullpath
			break
		}
	}

	if c.path == "" {
		var pathErrors []string

		for _, dir := range configPaths {
			if err := os.Mkdir(dir.fullpath, os.ModePerm); err == nil {
				c.path = dir.fullpath
				break
			}

			pathErrors = append(pathErrors, dir.fullpath)
		}

		if len(pathErrors) == len(configPaths) {
			dirError := "The configuration directories could not be created at:\n"
			dirError += strings.Join(pathErrors, "\n")

			PrintError(dirError)
		}
	}

	return
}

// ConfigPath returns the absolute path for the given configType.
func ConfigPath(configType string) (string, error) {
	confPath := filepath.Join(config.path, configType)

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
func GetProperty(property string) string {
	return config.String(property)
}

// IsPropertyEnabled returns if a property is enabled.
func IsPropertyEnabled(property string) bool {
	return config.Bool(property)
}

// AddConfigProperty adds a property and its value to the properties store.
func AddProperty(property string, value interface{}) {
	config.Set(property, value)
}

// generate generates and updates the configuration.
// Any existing values are appended to it.
func generate() {
	parseOldConfig()

	genMap := make(map[string]interface{})

	for _, option := range options {
		if !option.IsBoolean {
			genMap[option.Name] = config.Get(option.Name)
		}
	}

	keys := config.Get("keybindings")
	if keys == nil {
		keys = make(map[string]interface{})
	}
	genMap["keybindings"] = keys

	theme := config.Get("theme")
	if t, ok := theme.(string); ok && t == "" {
		theme = make(map[string]interface{})
	}
	genMap["theme"] = theme

	data, err := hjson.Marshal(genMap)
	if err != nil {
		PrintError(err.Error())
	}

	conf, err := ConfigPath("bluetuith.conf")
	if err != nil {
		PrintError(err.Error())
	}

	file, err := os.OpenFile(conf, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		PrintError(err.Error())
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		PrintError(err.Error())
	}

	if err := file.Sync(); err != nil {
		PrintError(err.Error())
	}
}

// parseOldConfig parses and stores values from the old configuration.
func parseOldConfig() {
	file, err := ConfigPath("config")
	if err != nil {
		return
	}

	fd, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		PrintWarn("Config: The old configuration could not be read")
		return
	}

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		values := strings.Split(line, "=")
		if len(values) != 2 {
			continue
		}

		AddProperty(values[0], values[1])
	}

	fd.Close()

	if err = scanner.Err(); err != nil && err != io.EOF {
		PrintWarn("Config: The old configuration could not be parsed")
	}
}
