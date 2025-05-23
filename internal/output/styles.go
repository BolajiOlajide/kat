/*
Styling using lipgloss library for better terminal styling
*/
package output

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strconv"
)

type Style interface {
	String() string
	Render(content string) string
}

// StyleText wraps a string with a style
func StyleText(s Style, text string) string {
	return s.Render(text)
}

// Info styles text with the info style
func Info(text string) string {
	return StyleText(StyleInfo, text)
}

// Infof formats and styles text with the info style
func Infof(format string, args ...interface{}) string {
	return StyleText(StyleInfo, fmt.Sprintf(format, args...))
}

// Success styles text with the success style
func Success(text string) string {
	return StyleText(StyleSuccess, text)
}

// Successf formats and styles text with the success style
func Successf(format string, args ...interface{}) string {
	return StyleText(StyleSuccess, fmt.Sprintf(format, args...))
}

// Warning styles text with the warning style
func Warning(text string) string {
	return StyleText(StyleWarning, text)
}

// Warningf formats and styles text with the warning style
func Warningf(format string, args ...interface{}) string {
	return StyleText(StyleWarning, fmt.Sprintf(format, args...))
}

// Failure styles text with the failure style
func Failure(text string) string {
	return StyleText(StyleFailure, text)
}

// Failuref formats and styles text with the failure style
func Failuref(format string, args ...interface{}) string {
	return StyleText(StyleFailure, fmt.Sprintf(format, args...))
}

// Heading styles text with the heading style
func Heading(text string) string {
	return StyleText(StyleHeading, text)
}

// Headingf formats and styles text with the heading style
func Headingf(format string, args ...interface{}) string {
	return StyleText(StyleHeading, fmt.Sprintf(format, args...))
}

// LipglossStyle wraps lipgloss.Style to implement our Style interface
type LipglossStyle struct {
	style lipgloss.Style
}

func (s *LipglossStyle) String() string {
	return s.style.String()
}

func (s *LipglossStyle) Render(content string) string {
	return s.style.Render(content)
}

// NewStyle creates a new style with lipgloss
func NewStyle(style lipgloss.Style) Style {
	return &LipglossStyle{style: style}
}

// CombineStyles combines multiple styles into a single style.
func CombineStyles(styles ...Style) Style {
	base := lipgloss.NewStyle()
	for _, s := range styles {
		if lipStyle, ok := s.(*LipglossStyle); ok {
			base = base.Inherit(lipStyle.style)
		}
	}
	return &LipglossStyle{style: base}
}

// Fg256Color returns a style that sets the foreground color to the given 256-color code.
func Fg256Color(code int) Style {
	return &LipglossStyle{style: lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(code)))}
}

// Bg256Color returns a style that sets the background color to the given 256-color code.
func Bg256Color(code int) Style {
	return &LipglossStyle{style: lipgloss.NewStyle().Background(lipgloss.Color(strconv.Itoa(code)))}
}

var (
	// General styles.

	// StyleReset is a style that resets all styles.
	// It should be used at the end of every styled string.
	StyleReset = &LipglossStyle{style: lipgloss.NewStyle()}

	StylePending = Fg256Color(4)
	// StyleInfo is a style that is used for informational messages.
	StyleInfo       = StylePending
	StyleWarning    = Fg256Color(124)
	StyleFailure    = NewStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true))
	StyleSuccess    = Fg256Color(2)
	StyleSuggestion = Fg256Color(244)
	StyleHeading    = NewStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true))

	StyleBold      = NewStyle(lipgloss.NewStyle().Bold(true))
	StyleItalic    = NewStyle(lipgloss.NewStyle().Italic(true))
	StyleUnderline = NewStyle(lipgloss.NewStyle().Underline(true))

	StyleWhiteOnPurple  = NewStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("55")))
	StyleGreyBackground = NewStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("242")))

	StyleLinesDeleted = Fg256Color(196)
	StyleLinesAdded   = Fg256Color(2)

	// Colors
	StyleGrey   = Fg256Color(7)
	StyleYellow = Fg256Color(220)
	StyleOrange = Fg256Color(202)
)
