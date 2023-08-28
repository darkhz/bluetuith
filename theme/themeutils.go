package theme

import (
	"fmt"
	"sort"

	"github.com/alexeyco/simpletable"
	"github.com/gdamore/tcell/v2"
)

// ColorWrap wraps the text content with the modifier element's color.
func ColorWrap(elementName ThemeContext, elementContent string, attributes ...string) string {
	attr := "::b"
	if attributes != nil {
		attr = attributes[0]
	}

	return fmt.Sprintf("[%s%s]%s[-:-:-]", ThemeConfig[elementName], attr, elementContent)
}

// BackgroundColor checks whether the given color is a light
// or dark color, and returns the appropriate color that is
// visible on top of the given color.
func BackgroundColor(themeContext ThemeContext) tcell.Color {
	if isLightColor(GetColor(themeContext)) {
		return tcell.Color16
	}

	return tcell.ColorWhite
}

// GetColor returns the color of the modifier element.
func GetColor(themeContext ThemeContext) tcell.Color {
	color := ThemeConfig[themeContext]
	if color == "black" {
		return tcell.Color16
	}

	return tcell.GetColor(color)
}

// GetElementData returns the element types and colors in a tabular format.
func GetElementData() string {
	var elements, colors []string

	for element := range ThemeConfig {
		elements = append(elements, string(element))
	}
	sort.Strings(elements)

	for color := range tcell.ColorNames {
		colors = append(colors, color)
	}
	sort.Strings(colors)

	elementsTable := simpletable.New()
	elementsTable.SetStyle(simpletable.StyleUnicode)
	elementsTable.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Theme Element Types"},
			{Align: simpletable.AlignCenter, Text: "Theme Colors"},
		},
	}

	for i, e := range elements {
		r := []*simpletable.Cell{}
		r = append(r, &simpletable.Cell{Text: e}, &simpletable.Cell{Text: colors[i]})

		elementsTable.Body.Cells = append(elementsTable.Body.Cells, r)

	}

	for _, c := range colors[len(elements):] {
		r := []*simpletable.Cell{}
		r = append(r, &simpletable.Cell{Text: ""}, &simpletable.Cell{Text: c})

		elementsTable.Body.Cells = append(elementsTable.Body.Cells, r)
	}

	return elementsTable.String()
}

// isLightColor checks if the given color is a light color.
// Adapted from:
// https://github.com/bgrins/TinyColor/blob/master/tinycolor.js#L68
func isLightColor(color tcell.Color) bool {
	r, g, b := color.RGB()
	brightness := (r*299 + g*587 + b*114) / 1000

	return brightness > 130
}

// isValidElementColor returns whether the modifier-value pair is valid.
func isValidElementColor(color string) bool {
	if color == "transparent" ||
		tcell.GetColor(color) != tcell.ColorDefault {
		return true
	}

	return false
}
