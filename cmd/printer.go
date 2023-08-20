package cmd

import (
	"os"

	"github.com/fatih/color"
)

// Print displays a message.
func Print(message string, status ...int) {
	color.New(color.FgWhite, color.Bold).Println(message)

	if status != nil {
		os.Exit(status[0])
	}
}

// PrintError prints an error to the screen.
func PrintError(message string, err ...error) {
	message = "[!] " + message

	color.New(color.FgRed, color.Bold).Println(message)
	os.Exit(1)
}
