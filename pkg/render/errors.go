package render

import (
	"fmt"
	"html"
	"strings"
)

const (
	errorCardWidth  = 600
	errorCardHeight = 120
	errorLineMax    = 96
)

// errorSVGCard renders a valid inline <svg> describing a failed diagram. It is
// the default never-fail fallback: markdown conversion keeps working, and the
// reader sees what went wrong instead of a blank page or a Go error.
func errorSVGCard(lang string, line int, err error) string {
	message := "unknown error"
	if err != nil {
		message = err.Error()
	}
	where := lang
	if line > 0 {
		where = fmt.Sprintf("%s, line %d", lang, line)
	}
	head, rest := splitMessage(message)

	var b strings.Builder
	fmt.Fprintf(&b,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" class="svg-diagram-svg svg-diagram-error" role="img" aria-label="%s diagram error">`,
		errorCardWidth, errorCardHeight, html.EscapeString(lang))
	fmt.Fprintf(&b,
		`<rect x="0.5" y="0.5" width="%d" height="%d" rx="8" fill="var(--diagram-error-bg, #fef2f2)" stroke="var(--diagram-error-border, #fca5a5)"/>`,
		errorCardWidth-1, errorCardHeight-1)
	fmt.Fprintf(&b,
		`<text x="20" y="38" font-family="ui-monospace, SFMono-Regular, Menlo, monospace" font-size="15" font-weight="600" fill="var(--diagram-error-fg, #b91c1c)">%s</text>`,
		html.EscapeString("diagram failed: "+where))
	fmt.Fprintf(&b,
		`<text x="20" y="66" font-family="ui-monospace, SFMono-Regular, Menlo, monospace" font-size="13" fill="var(--diagram-fg, #1e293b)">%s</text>`,
		html.EscapeString(head))
	fmt.Fprintf(&b,
		`<text x="20" y="94" font-family="ui-monospace, SFMono-Regular, Menlo, monospace" font-size="12" fill="var(--diagram-muted, #64748b)">%s</text>`,
		html.EscapeString(rest))
	b.WriteString(`</svg>`)
	return b.String()
}

// splitMessage breaks an error message into two card-sized lines.
func splitMessage(message string) (head, rest string) {
	head = truncate(message, errorLineMax)
	return head, truncate(strings.TrimSpace(message[len(head):]), errorLineMax)
}

// codeBlockHTML is the ErrorCodeBlock fallback: the fence survives as an
// ordinary code block instead of a diagram.
func codeBlockHTML(lang, source string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<pre><code class="language-%s">`, html.EscapeString(lang))
	b.WriteString(html.EscapeString(ensureTrailingNewline(source)))
	b.WriteString(`</code></pre>`)
	return b.String()
}

func ensureTrailingNewline(s string) string {
	if s == "" || strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

// truncate shortens s to at most n bytes without splitting a rune.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8Start(s[cut]) {
		cut--
	}
	return s[:cut]
}

func utf8Start(b byte) bool { return b&0xC0 != 0x80 }
