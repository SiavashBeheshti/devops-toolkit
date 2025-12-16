package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
)

// TableConfig holds configuration for table rendering
type TableConfig struct {
	Title       string
	Headers     []string
	ShowBorder  bool
	ShowRowLine bool
	Colored     bool
	Compact     bool
}

// Table represents a styled table
type Table struct {
	config TableConfig
	rows   [][]string
	colors [][]tablewriter.Colors
}

// NewTable creates a new styled table
func NewTable(config TableConfig) *Table {
	return &Table{
		config: config,
		rows:   make([][]string, 0),
		colors: make([][]tablewriter.Colors, 0),
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
	t.colors = append(t.colors, nil)
}

// AddColoredRow adds a row with specific colors
func (t *Table) AddColoredRow(row []string, colors []tablewriter.Colors) {
	t.rows = append(t.rows, row)
	t.colors = append(t.colors, colors)
}

// Render renders the table to stdout
func (t *Table) Render() {
	t.RenderTo(os.Stdout)
}

// RenderTo renders the table to the specified writer
func (t *Table) RenderTo(w io.Writer) {
	// Print title if present
	if t.config.Title != "" {
		titleBox := HeaderBoxStyle.Render(t.config.Title)
		fmt.Fprintln(w, titleBox)
		fmt.Fprintln(w)
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader(t.config.Headers)

	// Styling
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetTablePadding("  ")

	if t.config.ShowBorder {
		table.SetBorder(true)
		table.SetCenterSeparator("┼")
		table.SetColumnSeparator("│")
		table.SetRowSeparator("─")
		table.SetHeaderLine(true)
		table.SetRowLine(false)
	} else {
		table.SetBorder(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("  ")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
	}

	if t.config.ShowRowLine {
		table.SetRowLine(true)
	}

	// Header colors
	headerColors := make([]tablewriter.Colors, len(t.config.Headers))
	for i := range headerColors {
		headerColors[i] = tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiMagentaColor}
	}
	table.SetHeaderColor(headerColors...)

	// Add rows
	for i, row := range t.rows {
		if t.colors[i] != nil {
			table.Rich(row, t.colors[i])
		} else {
			table.Append(row)
		}
	}

	table.Render()
}

// StatusTable creates a pre-configured status table
func StatusTable(title string) *Table {
	return NewTable(TableConfig{
		Title:      title,
		Headers:    []string{"Component", "Status", "Details"},
		ShowBorder: true,
		Colored:    true,
	})
}

// ResourceTable creates a pre-configured resource table
func ResourceTable(title string, headers []string) *Table {
	return NewTable(TableConfig{
		Title:      title,
		Headers:    headers,
		ShowBorder: true,
		Colored:    true,
	})
}

// SimpleTable creates a borderless simple table
func SimpleTable(headers []string) *Table {
	return NewTable(TableConfig{
		Headers:    headers,
		ShowBorder: false,
		Colored:    true,
	})
}

// StatusRow returns a colored row based on status
func StatusRow(component, status, details string) ([]string, []tablewriter.Colors) {
	row := []string{component, status, details}
	var colors []tablewriter.Colors

	statusLower := strings.ToLower(status)
	switch {
	case strings.Contains(statusLower, "ok") || strings.Contains(statusLower, "running") ||
		strings.Contains(statusLower, "success") || strings.Contains(statusLower, "healthy") ||
		strings.Contains(statusLower, "ready") || strings.Contains(statusLower, "passed"):
		colors = []tablewriter.Colors{
			{tablewriter.FgHiWhiteColor},
			{tablewriter.Bold, tablewriter.FgGreenColor},
			{tablewriter.FgHiBlackColor},
		}
	case strings.Contains(statusLower, "warn") || strings.Contains(statusLower, "pending") ||
		strings.Contains(statusLower, "degraded"):
		colors = []tablewriter.Colors{
			{tablewriter.FgHiWhiteColor},
			{tablewriter.Bold, tablewriter.FgYellowColor},
			{tablewriter.FgHiBlackColor},
		}
	case strings.Contains(statusLower, "error") || strings.Contains(statusLower, "fail") ||
		strings.Contains(statusLower, "crash") || strings.Contains(statusLower, "critical"):
		colors = []tablewriter.Colors{
			{tablewriter.FgHiWhiteColor},
			{tablewriter.Bold, tablewriter.FgRedColor},
			{tablewriter.FgHiBlackColor},
		}
	default:
		colors = []tablewriter.Colors{
			{tablewriter.FgHiWhiteColor},
			{tablewriter.FgCyanColor},
			{tablewriter.FgHiBlackColor},
		}
	}

	return row, colors
}

// Panel renders a styled panel/box with content
func Panel(title, content string) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Width(60)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB"))

	fullContent := titleStyle.Render(title) + "\n\n" + contentStyle.Render(content)
	return style.Render(fullContent)
}

// KeyValue renders a key-value pair
func KeyValue(key, value string) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		Width(20)
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB"))

	return keyStyle.Render(key+":") + " " + valueStyle.Render(value)
}

// Divider renders a horizontal divider
func Divider(width int) string {
	return MutedStyle.Render(strings.Repeat("─", width))
}

// Section renders a section header
func Section(title string) string {
	return "\n" + TitleStyle.Render("▸ "+title) + "\n"
}

// SubSection renders a subsection header
func SubSection(title string) string {
	return SubtitleStyle.Render("  "+IconBullet+" "+title) + "\n"
}

