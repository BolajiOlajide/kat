package output

import (
	"fmt"
	"io"
	"os"
)

// Output manages the output of styled text to a writer
type Output struct {
	Writer io.Writer
}

// New creates a new Output that writes to stdout
func New() *Output {
	return &Output{Writer: os.Stdout}
}

// NewWithWriter creates a new Output with a custom writer
func NewWithWriter(w io.Writer) *Output {
	return &Output{Writer: w}
}

// Print prints a message with the default style
func (o *Output) Print(msg string) {
	fmt.Fprintln(o.Writer, msg)
}

// Printf prints a formatted message with the default style
func (o *Output) Printf(format string, args ...interface{}) {
	fmt.Fprintf(o.Writer, format+"\n", args...)
}

// Success prints a message with the success style
func (o *Output) Success(msg string) {
	fmt.Fprintln(o.Writer, Success(msg))
}

// Successf prints a formatted message with the success style
func (o *Output) Successf(format string, args ...interface{}) {
	fmt.Fprintln(o.Writer, Successf(format, args...))
}

// Info prints a message with the info style
func (o *Output) Info(msg string) {
	fmt.Fprintln(o.Writer, Info(msg))
}

// Infof prints a formatted message with the info style
func (o *Output) Infof(format string, args ...interface{}) {
	fmt.Fprintln(o.Writer, Infof(format, args...))
}

// Warning prints a message with the warning style
func (o *Output) Warning(msg string) {
	fmt.Fprintln(o.Writer, Warning(msg))
}

// Warningf prints a formatted message with the warning style
func (o *Output) Warningf(format string, args ...interface{}) {
	fmt.Fprintln(o.Writer, Warningf(format, args...))
}

// Failure prints a message with the failure style
func (o *Output) Failure(msg string) {
	fmt.Fprintln(o.Writer, Failure(msg))
}

// Failuref prints a formatted message with the failure style
func (o *Output) Failuref(format string, args ...interface{}) {
	fmt.Fprintln(o.Writer, Failuref(format, args...))
}

// Heading prints a message with the heading style
func (o *Output) Heading(msg string) {
	fmt.Fprintln(o.Writer, Heading(msg))
}

// Headingf prints a formatted message with the heading style
func (o *Output) Headingf(format string, args ...interface{}) {
	fmt.Fprintln(o.Writer, Headingf(format, args...))
}

// Default is a global output instance that writes to stdout
var Default = New()

// Methods with emoji

// SuccessEmoji prints a message with the success style and emoji
func (o *Output) SuccessEmoji(msg string) {
	o.Printf("%s %s", EmojiSuccess, Success(msg))
}

// SuccessEmojif formats and prints a message with the success style and emoji
func (o *Output) SuccessEmojif(format string, args ...interface{}) {
	o.Printf("%s %s", EmojiSuccess, Successf(format, args...))
}

// InfoEmoji prints a message with the info style and emoji
func (o *Output) InfoEmoji(msg string) {
	o.Printf("%s %s", EmojiInfo, Info(msg))
}

// InfoEmojif formats and prints a message with the info style and emoji
func (o *Output) InfoEmojif(format string, args ...interface{}) {
	o.Printf("%s %s", EmojiInfo, Infof(format, args...))
}

// WarningEmoji prints a message with the warning style and emoji
func (o *Output) WarningEmoji(msg string) {
	o.Printf("%s %s", EmojiWarning, Warning(msg))
}

// WarningEmojif formats and prints a message with the warning style and emoji
func (o *Output) WarningEmojif(format string, args ...interface{}) {
	o.Printf("%s %s", EmojiWarning, Warningf(format, args...))
}

// FailureEmoji prints a message with the failure style and emoji
func (o *Output) FailureEmoji(msg string) {
	o.Printf("%s %s", EmojiFailure, Failure(msg))
}

// FailureEmojif formats and prints a message with the failure style and emoji
func (o *Output) FailureEmojif(format string, args ...interface{}) {
	o.Printf("%s %s", EmojiFailure, Failuref(format, args...))
}