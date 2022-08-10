package theme

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ColorWrap wraps the text content with the modifier element's color.
func ColorWrap(elementName, elementContent string, attributes ...string) string {
	attr := "::b"
	if attributes != nil {
		attr = "::" + attributes[0]
	}

	return fmt.Sprintf("[%s%s]%s", ThemeConfig[elementName], attr, elementContent)
}

// BackgroundColor checks whether the given color is a light
// or dark color, and returns the appropriate color that is
// visible on top of the given color.
func BackgroundColor(textColor string) tcell.Color {
	if isLightColor(GetColor(textColor)) {
		return tcell.Color16
	}

	return tcell.ColorWhite
}

// GetColor returns the color of the modifier element.
func GetColor(colorType string) tcell.Color {
	color := ThemeConfig[colorType]
	if color == "black" {
		return tcell.Color16
	}

	return tcell.GetColor(color)
}

// GetElementModifiers returns a list of modifiers.
func GetElementModifiers() string {
	var modifiers string
	var elements []string

	for element := range ThemeConfig {
		elements = append(elements, element)
	}
	sort.Strings(elements)

	for _, element := range elements {
		modifiers += element + ", \n"
	}

	return strings.TrimRight(modifiers, ", \n")
}

// GetElementColors returns a list of colors for the modifiers.
func GetElementColors() string {
	var count int
	var colors []string
	var modifierColors string

	for color := range tcell.ColorNames {
		colors = append(colors, color)
	}
	sort.Strings(colors)

	for _, color := range colors {
		if count > 60 {
			count = 0
			modifierColors += "\n"
		}

		text := color + ", "
		modifierColors += text
		count += len(text)
	}

	return strings.TrimRight(modifierColors, ", ")
}

// GetThemes returns a list of themes that are stored
// in the themes directory.
func GetThemes() string {
	var count int
	var themeFiles string

	files, err := os.ReadDir(themesDir)
	if err != nil {
		return ""
	}

	for _, file := range files {
		if count > 60 {
			count = 0
			themeFiles += "\n"
		}

		text := file.Name() + ", "
		themeFiles += text
		count += len(themeFiles)
	}

	if themeFiles != "" {
		themeFiles = "\n\nAvailable themes:\n" + themeFiles
	}

	return strings.TrimRight(themeFiles, ", ")
}

// isLightColor checks if the given color is a light color.
// Adapted from:
// https://github.com/bgrins/TinyColor/blob/master/tinycolor.js#L68
func isLightColor(color tcell.Color) bool {
	r, g, b := color.RGB()
	brightness := (r*299 + g*587 + b*114) / 1000

	return brightness > 128
}

// isValidElementColor returns whether the modifier-value pair is valid.
func isValidElementColor(elementWithColor []string) bool {
	if _, ok := ThemeConfig[elementWithColor[0]]; ok {
		if elementWithColor[1] == "transparent" ||
			tcell.GetColor(elementWithColor[1]) != tcell.ColorDefault {
			return true
		}
	}

	return false
}
