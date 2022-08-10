package theme

import (
	"errors"
	"strings"
)

// ThemeConfig stores a list of color for the modifier elements.
var ThemeConfig = map[string]string{
	"Text":        "white",
	"Border":      "white",
	"Background":  "default",
	"StatusInfo":  "white",
	"StatusError": "red",

	"Adapter": "white",

	"Device":                   "white",
	"DeviceType":               "white",
	"DeviceConnected":          "white",
	"DeviceDiscovered":         "white",
	"DeviceProperty":           "grey",
	"DevicePropertyConnected":  "green",
	"DevicePropertyDiscovered": "orange",

	"Menu":     "white",
	"MenuBar":  "default",
	"MenuItem": "white",

	"ProgressBar":  "white",
	"ProgressText": "white",
}

// ParseThemeConfig parses the theme configuration.
func ParseThemeConfig(themeConfig string) error {
	elementsAndColors := strings.Split(themeConfig, ",")

	for _, elementAndColor := range elementsAndColors {
		elementWithColor := strings.Split(elementAndColor, "=")
		if len(elementWithColor) != 2 || !isValidElementColor(elementWithColor) {
			return errors.New(elementAndColor + ": Theme configuration is incorrect.")
		}

		switch elementWithColor[1] {
		case "black":
			elementWithColor[1] = "#000000"

		case "transparent":
			elementWithColor[1] = "default"
		}

		ThemeConfig[elementWithColor[0]] = elementWithColor[1]
	}

	return nil
}
