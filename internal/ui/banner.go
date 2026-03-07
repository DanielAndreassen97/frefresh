package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var bannerStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("6")).
	Padding(0, 1).
	Width(44)

var titleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("6")).
	Bold(true)

var dimStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("8"))

const logo = `  __               _          _
 / _|_ _ ___ / _|_ _ ___ __| |_
|  _| '_/ -_)  _| '_/ -_|_-< ' \
|_| |_| \___|_| |_| \___/__/_||_|`

func Banner() string {
	content := titleStyle.Render(logo) + "\n\n  " + dimStyle.Render("Semantic Model Table Refresh")
	return bannerStyle.Render(content)
}
