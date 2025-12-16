package output

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
)

// Printer handles all CLI output
type Printer struct {
	spinner *spinner.Spinner
}

// NewPrinter creates a new Printer instance
func NewPrinter() *Printer {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Color("magenta", "bold")
	return &Printer{
		spinner: s,
	}
}

// Default printer instance
var defaultPrinter = NewPrinter()

// Print outputs a message
func Print(msg string) {
	fmt.Println(msg)
}

// Printf outputs a formatted message
func Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Success prints a success message
func Success(msg string) {
	icon := SuccessStyle.Render(IconSuccess)
	fmt.Printf("%s %s\n", icon, msg)
}

// Successf prints a formatted success message
func Successf(format string, args ...interface{}) {
	Success(fmt.Sprintf(format, args...))
}

// Warning prints a warning message
func Warning(msg string) {
	icon := WarningStyle.Render(IconWarning)
	fmt.Printf("%s %s\n", icon, msg)
}

// Warningf prints a formatted warning message
func Warningf(format string, args ...interface{}) {
	Warning(fmt.Sprintf(format, args...))
}

// Error prints an error message
func Error(msg string) {
	icon := ErrorStyle.Render(IconError)
	fmt.Fprintf(os.Stderr, "%s %s\n", icon, msg)
}

// Errorf prints a formatted error message
func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

// Info prints an info message
func Info(msg string) {
	icon := InfoStyle.Render(IconInfo)
	fmt.Printf("%s %s\n", icon, msg)
}

// Infof prints a formatted info message
func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

// Muted prints a muted/dim message
func Muted(msg string) {
	fmt.Println(MutedStyle.Render(msg))
}

// Title prints a title
func Title(msg string) {
	fmt.Println()
	fmt.Println(TitleStyle.Render(msg))
}

// Subtitle prints a subtitle
func Subtitle(msg string) {
	fmt.Println(SubtitleStyle.Render(msg))
}

// Header prints a header with box style
func Header(msg string) {
	fmt.Println()
	fmt.Println(HeaderBoxStyle.Render(msg))
	fmt.Println()
}

// Banner prints an application banner
func Banner(name, version, description string) {
	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginTop(1).
		MarginBottom(1)

	versionStyle := lipgloss.NewStyle().
		Foreground(MutedColor)

	descStyle := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Italic(true)

	fmt.Println(bannerStyle.Render(name) + " " + versionStyle.Render(version))
	fmt.Println(descStyle.Render(description))
	fmt.Println()
}

// StartSpinner starts a spinner with message
func StartSpinner(msg string) {
	defaultPrinter.spinner.Suffix = " " + msg
	defaultPrinter.spinner.Start()
}

// StopSpinner stops the spinner
func StopSpinner() {
	defaultPrinter.spinner.Stop()
}

// SpinnerSuccess stops spinner with success message
func SpinnerSuccess(msg string) {
	defaultPrinter.spinner.Stop()
	Success(msg)
}

// SpinnerError stops spinner with error message
func SpinnerError(msg string) {
	defaultPrinter.spinner.Stop()
	Error(msg)
}

// ProgressBar renders a simple progress bar
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		total = 1
	}
	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))
	empty := width - filled

	bar := ""
	if filled > 0 {
		bar += SuccessStyle.Render(repeatChar("█", filled))
	}
	if empty > 0 {
		bar += MutedStyle.Render(repeatChar("░", empty))
	}

	return fmt.Sprintf("%s %3.0f%%", bar, percentage*100)
}

func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}

// List prints a bulleted list
func List(items []string) {
	for _, item := range items {
		fmt.Printf("  %s %s\n", MutedStyle.Render(IconBullet), item)
	}
}

// NumberedList prints a numbered list
func NumberedList(items []string) {
	for i, item := range items {
		num := InfoStyle.Render(fmt.Sprintf("%2d.", i+1))
		fmt.Printf("  %s %s\n", num, item)
	}
}

// Tree prints items in a tree structure
func Tree(root string, children []string) {
	fmt.Printf("  %s\n", root)
	for i, child := range children {
		prefix := IconTee
		if i == len(children)-1 {
			prefix = IconCorner
		}
		fmt.Printf("  %s%s %s\n", MutedStyle.Render(prefix), MutedStyle.Render(IconDash), child)
	}
}

// Badge returns a styled badge
func Badge(text, badgeType string) string {
	switch badgeType {
	case "success":
		return SuccessBadge.Render(text)
	case "warning":
		return WarningBadge.Render(text)
	case "error":
		return ErrorBadge.Render(text)
	case "info":
		return InfoBadge.Render(text)
	default:
		return InfoBadge.Render(text)
	}
}

// StatusIcon returns an icon based on status
func StatusIcon(status string) string {
	switch status {
	case "success", "ok", "running", "healthy", "ready", "passed":
		return SuccessStyle.Render(IconSuccess)
	case "warning", "pending", "degraded":
		return WarningStyle.Render(IconWarning)
	case "error", "failed", "crash", "critical":
		return ErrorStyle.Render(IconError)
	default:
		return InfoStyle.Render(IconInfo)
	}
}

// Summary prints a summary box
func Summary(title string, items map[string]string) {
	fmt.Println()
	fmt.Println(HeaderBoxStyle.Render(title))
	fmt.Println()
	for key, value := range items {
		fmt.Println(KeyValue(key, value))
	}
	fmt.Println()
}

// Newline prints an empty line
func Newline() {
	fmt.Println()
}

