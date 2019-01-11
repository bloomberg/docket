package compose

import (
	"os"

	"github.com/fatih/color"
)

func trace(format string, a ...interface{}) {
	fprintfGreen := color.New(color.FgBlue).FprintfFunc()
	fprintfGreen(os.Stderr, "[docket] ")
	fprintfGreen(os.Stderr, format, a...)
}
