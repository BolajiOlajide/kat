package loggr

import (
	"fmt"

	"github.com/BolajiOlajide/kat/internal/output"
)

type defaultLogger struct{}

func NewDefault() Logger {
	return &defaultLogger{}
}

func (d *defaultLogger) print(style output.Style, msg string) {
	fmt.Printf("%s%s%s\n", style, msg, output.StyleReset)
}

func (d *defaultLogger) Debug(msg string) {
	d.print(output.StyleSuggestion, msg)
}

func (d *defaultLogger) Info(msg string) {
	d.print(output.StyleInfo, msg)
}

func (d *defaultLogger) Warn(msg string) {
	d.print(output.StyleWarning, msg)
}

func (d *defaultLogger) Error(msg string) {
	d.print(output.StyleFailure, msg)
}
