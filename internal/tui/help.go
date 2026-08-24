package tui

import "strings"

func (m Model) writeHelp(b *strings.Builder) {
	b.WriteString(helpFooterTop)
	b.WriteString(helpFooterRest)
}
