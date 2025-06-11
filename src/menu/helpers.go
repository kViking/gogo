package menu

import (
	"gogo/src/style"
)

// MenuNavHelp returns the standard navigation help string for the menu (with Esc/q unified).
func MenuNavHelp() string {
	return style.MenuHelpStyle.Render("←/→ or Enter: Move, Esc/q: Back, confirm quit at top level")
}

// MenuQuitConfirm returns the quit confirmation prompt for the top level.
func MenuQuitConfirm() string {
	return style.ConfirmStyle.Render("Are you sure you want to quit GoGoGadget? (y/n)")
}

// EndpointNavHelp returns the standard navigation help for endpoint panels (info, confirm, etc).
func EndpointNavHelp() string {
	return style.MenuHelpStyle.Render("Esc/q: Back/Close")
}

// RenderMenuItem returns a styled menu item string based on selection and active state.
func RenderMenuItem(label string, selected, active bool) string {
	cursor := "  "
	if selected {
		cursor = "> "
	}
	if selected && active {
		return style.MenuSelectedStyle.Render(cursor + label)
	} else if active {
		return style.MenuActiveStyle.Render(cursor + label)
	}
	return style.MenuDimmedStyle.Render(cursor + label)
}

// ConfirmPrompt returns a styled confirmation prompt string.
func ConfirmPrompt(msg string) string {
	return style.HeaderStyle.Render(msg) + "\nAre you sure? (y/n)"
}
