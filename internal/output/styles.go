/*
This file was referenced from https://github.com/sourcegraph/sourcegraph-public-snapshot/blob/0aa8310ca4d924c084a7a3f836d7a4cbc45c3338/lib/output/style.go#L12
*/
package output

import (
	"fmt"
	"strings"
)

type Style interface {
	fmt.Stringer
}

// CombineStyles combines multiple styles into a single style.
func CombineStyles(styles ...Style) Style {
	sb := strings.Builder{}
	for _, s := range styles {
		fmt.Fprint(&sb, s)
	}
	return &style{sb.String()}
}

// Fg256Color returns a style that sets the foreground color to the given 256-color code.
func Fg256Color(code int) Style { return &style{fmt.Sprintf("\033[38;5;%dm", code)} }

// Bg256Color returns a style that sets the background color to the given 256-color code.
func Bg256Color(code int) Style { return &style{fmt.Sprintf("\033[48;5;%dm", code)} }

type style struct{ code string }

func (s *style) String() string { return s.code }

var (
	// General styles.

	// StyleReset is a style that resets all styles.
	// It should be used at the end of every styled string.
	StyleReset = &style{"\033[0m"}

	StylePending = Fg256Color(4)
	// StyleInfo is a style that is used for informational messages.
	StyleInfo       = StylePending
	StyleWarning    = Fg256Color(124)
	StyleFailure    = CombineStyles(StyleBold, Fg256Color(196))
	StyleSuccess    = Fg256Color(2)
	StyleSuggestion = Fg256Color(244)
	StyleHeading    = CombineStyles(StyleBold, Fg256Color(6))

	StyleBold      = &style{"\033[1m"}
	StyleItalic    = &style{"\033[3m"}
	StyleUnderline = &style{"\033[4m"}

	StyleWhiteOnPurple  = CombineStyles(Fg256Color(255), Bg256Color(55))
	StyleGreyBackground = CombineStyles(Fg256Color(0), Bg256Color(242))

	StyleLinesDeleted = Fg256Color(196)
	StyleLinesAdded   = Fg256Color(2)

	// Colors
	StyleGrey   = Fg256Color(7)
	StyleYellow = Fg256Color(220)
	StyleOrange = Fg256Color(202)
)
