// cliui.go: Centralized CLI UI utilities (styled output, spinner, etc.)
package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/muesli/termenv"
)

var termProfile = termenv.ColorProfile()

// Miami Nights color scheme (official palette, only needed variables)
var miamiNightsColors = map[string]string{
	"error":      "#FF5370", // pinkish red
	"success":    "#C3E88D", // light green
	"warning":    "#FFCB6B", // yellow/orange
	"info":       "#82AAFF", // blue
	"title":      "#FFFFFF", // white
	"variable":   "#C792EA", // purple
	"highlight":  "#F78C6C", // orange
	"gogogadget": "#FFD600", // bright yellow (distinct from warning)
}

// Monokai color scheme (classic palette, only needed variables)
var monokaiColors = map[string]string{
	"error":      "#F92672", // pink/red
	"success":    "#A6E22E", // green
	"warning":    "#FD971F", // orange
	"info":       "#66D9EF", // cyan/blue
	"title":      "#F8F8F2", // near white
	"variable":   "#AE81FF", // purple
	"highlight":  "#E6DB74", // yellow
	"gogogadget": "#FFD600", // bright yellow (distinct from warning)
}

// getActiveColorScheme loads the user's color scheme choice from settings.json
func getActiveColorScheme() map[string]string {
	settingsPath := "settings.json"
	file, err := os.Open(settingsPath)
	if err != nil {
		return MiamiNights // fallback default
	}
	defer file.Close()
	var settings struct {
		ColorScheme string `json:"colorScheme"`
	}
	if err := json.NewDecoder(file).Decode(&settings); err != nil {
		return MiamiNights
	}
	switch settings.ColorScheme {
	case "miaminights":
		return MiamiNights
	case "miaminights-light":
		return MiamiNightsLight
	case "monokai":
		return Monokai
	case "monokai-light":
		return MonokaiLight
	case "darcula":
		return Darcula
	case "darcula-light":
		return DarculaLight
	case "solarized-dark":
		return SolarizedDark
	case "solarized-light":
		return SolarizedLight
	case "powershell-dark":
		return PowerShellDark
	case "powershell-light":
		return PowerShellLight
	default:
		return MiamiNights
	}
}

var activeColors = getActiveColorScheme()

// Spinner utility with rainbow animation using termenv
func RainbowSpinner(message string, duration time.Duration) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	colors := []string{"#FF5555", "#FFB86C", "#F1FA8C", "#50FA7B", "#8BE9FD", "#BD93F9", "#FF79C6"}
	start := time.Now()
	for time.Since(start) < duration {
		for i, frame := range frames {
			color := termProfile.Color(colors[i%len(colors)])
			styled := termenv.String(frame).Foreground(color)
			fmt.Printf("\r%s %s", styled, message)
			time.Sleep(80 * time.Millisecond)
			if time.Since(start) >= duration {
				break
			}
		}
	}
	fmt.Printf("\r✔ %s\n", message)
}

// Style keywords for semantic coloring using termenv
var styleMap = map[string]func(string) string{
	"error": func(s string) string {
		return termenv.String(s).Foreground(termProfile.Color(activeColors["error"])).String()
	},
	"success": func(s string) string {
		return termenv.String(s).Foreground(termProfile.Color(activeColors["success"])).String()
	},
	"warning": func(s string) string {
		return termenv.String(s).Foreground(termProfile.Color(activeColors["warning"])).String()
	},
	"info": func(s string) string {
		return termenv.String(s).Foreground(termProfile.Color(activeColors["info"])).String()
	},
	"title": func(s string) string {
		return termenv.String(s).Foreground(termProfile.Color(activeColors["title"])).String()
	},
	"variable": func(s string) string {
		return termenv.String(s).Foreground(termProfile.Color(activeColors["variable"])).String()
	},
	"highlight": func(s string) string {
		return termenv.String(s).Foreground(termProfile.Color(activeColors["highlight"])).String()
	},
	"gogogadget": func(s string) string {
		return termenv.String(s).Foreground(termProfile.Color(activeColors["gogogadget"])).String()
	},
	"reset": func(s string) string { return s },
}

type StyledChunk struct {
	Text  string
	Style string // e.g. "error", "title", "variable", etc.
}

type ColorText struct{}

var colorText = ColorText{}

func (ColorText) PrintStyledLine(chunks ...StyledChunk) {
	for _, chunk := range chunks {
		if styleFn, ok := styleMap[chunk.Style]; ok {
			fmt.Print(styleFn(chunk.Text))
		} else {
			fmt.Print(chunk.Text)
		}
	}
	fmt.Println()
}

// Convenience single-style helpers
func (ColorText) Style(style, msg string, a ...interface{}) {
	if styleFn, ok := styleMap[style]; ok {
		fmt.Println(styleFn(fmt.Sprintf(msg, a...)))
	} else {
		fmt.Printf(msg+"\n", a...)
	}
}

func (ColorText) Red(msg string)    { fmt.Println(styleMap["error"](msg)) }
func (ColorText) Green(msg string)  { fmt.Println(styleMap["success"](msg)) }
func (ColorText) Yellow(msg string) { fmt.Println(styleMap["warning"](msg)) }
func (ColorText) Cyan(msg string)   { fmt.Println(styleMap["info"](msg)) }
func (ColorText) Bold(msg string, a ...interface{}) {
	fmt.Println(termenv.String(fmt.Sprintf(msg, a...)).String())
}
func (ColorText) Magenta(msg string, a ...interface{}) {
	fmt.Println(termenv.String(fmt.Sprintf(msg, a...)).Foreground(termProfile.Color("#BD93F9")).String())
}
