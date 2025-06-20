package scripts

import (
	"fmt"

	"github.com/mattn/go-colorable"
)

type ColorText struct{}

var colorText = ColorText{}

func (ColorText) Red(msg string) {
	fmt.Fprintln(colorable.NewColorableStderr(), "\x1b[31m"+msg+"\x1b[0m")
}

func (ColorText) Green(msg string) {
	fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[32m"+msg+"\x1b[0m")
}

func (ColorText) Yellow(msg string) {
	fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[33m"+msg+"\x1b[0m")
}

func (ColorText) Cyan(msg string) {
	fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[36m"+msg+"\x1b[0m")
}
