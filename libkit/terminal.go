package main

import (
	"fmt"
	"os"
	"strings"
)

var noColor = os.Getenv("NO_COLOR") != ""

const (
	reset      = "\x1b[0m"
	bold       = "\x1b[1m"
	dim        = "\x1b[2m"
	cyan       = "\x1b[36m"
	green      = "\x1b[32m"
	yellow     = "\x1b[33m"
	red        = "\x1b[31m"
	magenta    = "\x1b[35m"
	blue       = "\x1b[34m"
	brightCyan = "\x1b[96m"
)

func colorize(code, text string) string {
	if noColor || text == "" {
		return text
	}
	return code + text + reset
}

func printBanner() {
	banner := `
  _ _ _     _    _ _   
 | (_) |__ | | _(_) |_ 
 | | | '_ \| |/ / | __|
 | | | |_) |   <| | |_ 
 |_|_|_.__/|_|\_\_|\__|
  Unified Go Library Package Manager & Scaffold Tool
`
	fmt.Println(colorize(brightCyan, banner))
}

func printSuccess(msg string) {
	fmt.Printf("%s %s\n", colorize(green, "✔"), msg)
}

func printInfo(msg string) {
	fmt.Printf("%s %s\n", colorize(cyan, "ℹ"), msg)
}

func printWarn(msg string) {
	fmt.Printf("%s %s\n", colorize(yellow, "▲"), msg)
}

func printError(msg string) {
	fmt.Printf("%s %s\n", colorize(red, "✖"), msg)
}

func printTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("No records found.")
		return
	}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print header
	var headerLine strings.Builder
	for i, h := range headers {
		headerLine.WriteString(fmt.Sprintf("%-*s  ", colWidths[i], h))
	}
	fmt.Println(colorize(bold+cyan, headerLine.String()))

	// Print separator
	var sepLine strings.Builder
	for _, w := range colWidths {
		sepLine.WriteString(strings.Repeat("-", w) + "  ")
	}
	fmt.Println(colorize(dim, sepLine.String()))

	// Print rows
	for _, row := range rows {
		var rowLine strings.Builder
		for i, cell := range row {
			if i == 0 {
				rowLine.WriteString(fmt.Sprintf("%-*s  ", colWidths[i], cell))
			} else if i == 1 {
				rowLine.WriteString(colorize(bold+green, fmt.Sprintf("%-*s  ", colWidths[i], cell)))
			} else {
				rowLine.WriteString(fmt.Sprintf("%-*s  ", colWidths[i], cell))
			}
		}
		fmt.Println(rowLine.String())
	}
}
