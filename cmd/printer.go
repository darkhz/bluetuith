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

// PrintWarn prints a warning to the screen.
func PrintWarn(message string) {
	message = "[-] " + message

	color.New(color.FgYellow, color.Bold).Println(message)
}

// PrintError prints an error to the screen.
func PrintError(message string, err ...error) {
	message = "[!] " + message

	color.New(color.FgRed, color.Bold).Println(message)
	os.Exit(1)
}
