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

var config Config

// setup checks for the config directory,
// and creates one if it doesn't exist.
func (c *Config) setup() {
	var configExists bool

	c.Koanf = koanf.New(".")

	homedir, err := os.UserHomeDir()
	if err != nil {
		PrintError(err.Error())
	}

	dirs := []string{".config/bluetuith", ".bluetuith"}
	for i, dir := range dirs {
		p := filepath.Join(homedir, dir)
		dirs[i] = p

		if _, err := os.Stat(p); err == nil {
			c.path = p
			return
		}

		if i > 0 {
			continue
		}

		if _, err := os.Stat(filepath.Clean(filepath.Dir(p))); err == nil {
			configExists = true
		}
	}

	if c.path == "" {
		var pos int
		var err error

		if configExists {
			err = os.Mkdir(dirs[0], 0700)
		} else {
			pos = 1
			err = os.Mkdir(dirs[1], 0700)
		}

		if err != nil {
			PrintError(err.Error())
		}

		c.path = dirs[pos]
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
