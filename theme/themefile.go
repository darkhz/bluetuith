package theme

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var themesDir string

// CreateThemesDir creates a directory to store themes
// in the user's config path.
func CreateThemesDir(configPath string) error {
	themesDir = filepath.Join(configPath, "themes")

	if _, err := os.Stat(themesDir); err == nil {
		return nil
	}

	return os.Mkdir(themesDir, 0700)
}

// ParseThemeFile parses the theme files which are stored
// in the themes directory.
func ParseThemeFile(theme string) error {
	var themeConfig string

	file, err := os.Open(filepath.Join(themesDir, theme))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		if line[:1] == "#" {
			continue
		}

		for i, v := range line {
			if v == '=' {
				themeConfig += line[:i] + "=" + line[i+1:] + ","
			}
		}
	}

	return ParseThemeConfig(strings.TrimRight(themeConfig, ","))
}
