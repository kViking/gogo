// spinner.go: Provides a conventional, reusable spinner for CLI feedback.
package scripts

import (
	"time"

	"github.com/briandowns/spinner"
)

// NewSpinner returns a new spinner instance with a standard style and message.
func NewSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // CharSets[14] is a dots style
	s.Suffix = " " + message
	return s
}
