package scripts

import (
	"time"

	"github.com/briandowns/spinner"
)

// GetSpinner returns a new spinner instance with a standard style and message.
func GetSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // CharSets[14] is a pretty dots style
	s.Suffix = " " + message
	return s
}
