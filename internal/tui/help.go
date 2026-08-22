package tui

import "strings"

func (m Model) writeHelp(b *strings.Builder) {
	b.WriteString("\n↑↓ Select (u/j)   Enter Details   / Filter   K Kill\n")
	b.WriteString("R Refresh   V Views   ? Help   Q Quit\n")
}
