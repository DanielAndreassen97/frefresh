package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var dimStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("8"))

// Version is set by main at startup.
var Version = "dev"

// Each letter: 8 rows, 7 columns
var pixelFont = map[rune][8]string{
	'F': {
		"1111110",
		"1111110",
		"1100000",
		"1111100",
		"1111100",
		"1100000",
		"1100000",
		"1100000",
	},
	'R': {
		"1111100",
		"1100110",
		"1100110",
		"1111100",
		"1100110",
		"1100110",
		"1100110",
		"1100110",
	},
	'E': {
		"1111110",
		"1111110",
		"1100000",
		"1111100",
		"1111100",
		"1100000",
		"1111110",
		"1111110",
	},
	'S': {
		"0111110",
		"1100110",
		"1100000",
		"0111100",
		"0011110",
		"0000110",
		"1100110",
		"0111100",
	},
	'H': {
		"1100110",
		"1100110",
		"1100110",
		"1111110",
		"1111110",
		"1100110",
		"1100110",
		"1100110",
	},
}

const (
	letterW  = 7
	letterH  = 8
	spacing  = 1
	shadowDx = 1
	shadowDy = 1
)

func hslToRGB(h, s, l float64) (int, int, int) {
	c := (1 - math.Abs(2*l-1)) * s
	hh := h / 60.0
	x := c * (1 - math.Abs(math.Mod(hh, 2)-1))
	var r, g, b float64
	switch {
	case hh < 1:
		r, g, b = c, x, 0
	case hh < 2:
		r, g, b = x, c, 0
	case hh < 3:
		r, g, b = 0, c, x
	case hh < 4:
		r, g, b = 0, x, c
	case hh < 5:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	m := l - c/2
	return int((r + m) * 255), int((g + m) * 255), int((b + m) * 255)
}

func buildGrid(word string) ([][]bool, int, int) {
	w := len(word)*(letterW+spacing) - spacing
	h := letterH
	grid := make([][]bool, h)
	for r := range grid {
		grid[r] = make([]bool, w)
	}
	for i, ch := range word {
		pix, ok := pixelFont[ch]
		if !ok {
			continue
		}
		xOff := i * (letterW + spacing)
		for row := 0; row < letterH; row++ {
			for col := 0; col < letterW && col < len(pix[row]); col++ {
				if pix[row][col] == '1' {
					grid[row][xOff+col] = true
				}
			}
		}
	}
	return grid, w, h
}

func gradientBanner() string {
	word := "FREFRESH"
	grid, w, h := buildGrid(word)

	shadowW := w + shadowDx
	shadowH := h + shadowDy
	shadow := make([][]bool, shadowH)
	for r := range shadow {
		shadow[r] = make([]bool, shadowW)
	}
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			if grid[r][c] {
				sr, sc := r+shadowDy, c+shadowDx
				if sr < shadowH && sc < shadowW {
					shadow[sr][sc] = true
				}
			}
		}
	}

	fullGrid := make([][]bool, shadowH)
	for r := range fullGrid {
		fullGrid[r] = make([]bool, shadowW)
	}
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			fullGrid[r][c] = grid[r][c]
		}
	}

	totalW := shadowW

	var sb strings.Builder
	sb.WriteString(" ")
	outputRows := (shadowH + 1) / 2
	for pair := 0; pair < outputRows; pair++ {
		topIdx := pair * 2
		botIdx := pair*2 + 1
		for col := 0; col < totalW; col++ {
			topMain := topIdx < shadowH && fullGrid[topIdx][col]
			botMain := botIdx < shadowH && fullGrid[botIdx][col]
			topShad := topIdx < shadowH && shadow[topIdx][col] && !topMain
			botShad := botIdx < shadowH && shadow[botIdx][col] && !botMain

			t := float64(col) / float64(totalW-1)
			hue := 12.0 + t*38.0
			r, g, b := hslToRGB(hue, 0.90, 0.55)
			mainColor := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))

			sr, sg, sbb := hslToRGB(hue, 0.3, 0.20)
			shadColor := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", sr, sg, sbb))

			if topMain && botMain {
				sb.WriteString(lipgloss.NewStyle().Foreground(mainColor).Render("█"))
			} else if topMain && botShad {
				sb.WriteString(lipgloss.NewStyle().Foreground(mainColor).Background(shadColor).Render("▀"))
			} else if topMain {
				sb.WriteString(lipgloss.NewStyle().Foreground(mainColor).Render("▀"))
			} else if topShad && botMain {
				sb.WriteString(lipgloss.NewStyle().Foreground(mainColor).Background(shadColor).Render("▄"))
			} else if botMain {
				sb.WriteString(lipgloss.NewStyle().Foreground(mainColor).Render("▄"))
			} else if topShad && botShad {
				sb.WriteString(lipgloss.NewStyle().Foreground(shadColor).Render("█"))
			} else if topShad {
				sb.WriteString(lipgloss.NewStyle().Foreground(shadColor).Render("▀"))
			} else if botShad {
				sb.WriteString(lipgloss.NewStyle().Foreground(shadColor).Render("▄"))
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n ")
	}
	return sb.String()
}

func gradientText(text string) string {
	runes := []rune(text)
	var sb strings.Builder
	for i, ch := range runes {
		t := float64(i) / float64(max(len(runes)-1, 1))
		hue := 12.0 + t*38.0
		r, g, b := hslToRGB(hue, 0.90, 0.55)
		color := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
		sb.WriteString(lipgloss.NewStyle().Foreground(color).Bold(true).Render(string(ch)))
	}
	return sb.String()
}

func centerPad(visibleLen, bannerWidth int) string {
	pad := (bannerWidth - visibleLen) / 2
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad)
}

func Banner() string {
	banner := gradientBanner()
	bannerWidth := len([]rune("FREFRESH"))*(letterW+spacing) - spacing + shadowDx

	verText := "v" + Version
	ver := gradientText(verText)

	hintText := "↑↓ navigate • 1-9 select • enter confirm • esc back • q quit"
	hint := dimStyle.Render(hintText)

	return "\n" + banner + "\n" +
		centerPad(len([]rune(verText)), bannerWidth) + ver + "\n" +
		centerPad(len([]rune(hintText)), bannerWidth) + hint
}
