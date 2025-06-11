package menu

import (
	scripts "gogo/src/scripts"
	"gogo/src/style"
)

type MenuColumn struct {
	Title    string
	Options  []string
	Selected int
	GetNext  func(selected string) *MenuColumn
}

func BuildRootMenu() *MenuColumn {
	return &MenuColumn{
		Title:    "GoGoGadget",
		Options:  []string{"Gadgets", "Add a new gadget", "Analyze a command", "Quit"},
		Selected: 0,
		GetNext: func(selected string) *MenuColumn {
			switch selected {
			case "Gadgets":
				gadgetNames := scripts.GetGadgetNames()
				if len(gadgetNames) == 0 {
					return &MenuColumn{Title: "Gadgets", Options: []string{"No gadgets found."}, Selected: 0}
				}
				return &MenuColumn{
					Title:    "Gadgets",
					Options:  gadgetNames,
					Selected: 0,
					GetNext: func(gadget string) *MenuColumn {
						return &MenuColumn{
							Title:    gadget,
							Options:  []string{"Run", "Show info", "Edit", "Delete"},
							Selected: 0,
							GetNext: func(subcmd string) *MenuColumn {
								return nil
							},
						}
					},
				}
			default:
				return nil
			}
		},
	}
}

// RenderColumn renders a menu column, including the navigation help message for the first column.
func RenderColumn(col *MenuColumn, active, dim bool) string {
	var s string
	if col.Title != "" {
		s += style.MenuHeaderStyle.Render(col.Title) + "\n"
	}
	for i, opt := range col.Options {
		cursor := "  "
		if col.Selected == i {
			cursor = "> "
		}
		if col.Selected == i && active {
			s += style.MenuSelectedStyle.Render(cursor+opt) + "\n"
		} else if active {
			s += style.MenuActiveStyle.Render(cursor+opt) + "\n"
		} else {
			s += style.MenuDimmedStyle.Render(cursor+opt) + "\n"
		}
	}
	// Add navigation help message at the bottom of the first column
	if active && col.Title == "GoGoGadget" {
		help := style.MenuHelpStyle.Render("←/→ or Enter: Move, Esc: Back, q: Quit")
		s += "\n" + help
	}
	return s
}
