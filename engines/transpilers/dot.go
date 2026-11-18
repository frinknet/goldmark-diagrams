// Package transpilers converts Graphviz DOT and PlantUML sources into the
// dialects consumed by the pure-Go renderers: D2 for most diagram forms, and
// Mermaid (rendered by pele-go) for the PlantUML gantt and mindmap forms.
package transpilers

import (
	"fmt"
	"strings"
)

// DotToD2 translates Graphviz DOT source into D2 source.
func DotToD2(src string) (string, error) {
	toks, err := dotLex(src)
	if err != nil {
		return "", err
	}
	p := &dotParser{
		toks:  toks,
		nodes: make(map[string]*dotNode),
	}
	if err := p.parse(); err != nil {
		return "", err
	}
	if len(p.out) == 0 {
		return "", fmt.Errorf("dot: no statements")
	}
	return strings.Join(p.out, "\n") + "\n", nil
}

// ---------------------------------------------------------------------------
// Lexer
// ---------------------------------------------------------------------------

type dotTokenKind int

const (
	dotEOF dotTokenKind = iota
	dotIdent
	dotString
	dotPunct
)

type dotToken struct {
	kind dotTokenKind
	text string
	line int
}

func dotIsSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\v' || c == '\f'
}

func dotIsPunct(c byte) bool {
	switch c {
	case '{', '}', '[', ']', ';', ',', '=', '"', ':':
		return true
	}
	return false
}

// dotLex splits DOT source into identifiers, quoted strings and punctuation,
// discarding //, # and /* */ comments.
func dotLex(src string) ([]dotToken, error) {
	var toks []dotToken
	line := 1
	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case c == '\n':
			line++
			i++
		case c == ' ' || c == '\t' || c == '\r' || c == '\v' || c == '\f':
			i++
		case c == '#':
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(src) && src[i+1] == '/':
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			start := line
			i += 2
			closed := false
			for i < len(src) {
				if src[i] == '*' && i+1 < len(src) && src[i+1] == '/' {
					i += 2
					closed = true
					break
				}
				if src[i] == '\n' {
					line++
				}
				i++
			}
			if !closed {
				return nil, fmt.Errorf("dot: unterminated comment starting at line %d", start)
			}
		case c == '"':
			start := line
			i++
			var b strings.Builder
			closed := false
			for i < len(src) {
				ch := src[i]
				if ch == '\\' && i+1 < len(src) {
					switch src[i+1] {
					case '"', '\\':
						b.WriteByte(src[i+1])
						i += 2
					case 'n', 'l', 'r':
						// DOT newline escapes become real newlines.
						b.WriteByte('\n')
						i += 2
					case '\n':
						line++
						i += 2
					default:
						b.WriteByte(ch)
						i++
					}
					continue
				}
				if ch == '"' {
					i++
					closed = true
					break
				}
				if ch == '\n' {
					line++
				}
				b.WriteByte(ch)
				i++
			}
			if !closed {
				return nil, fmt.Errorf("dot: unterminated quoted string starting at line %d", start)
			}
			toks = append(toks, dotToken{kind: dotString, text: b.String(), line: start})
		case c == '-' && i+1 < len(src) && src[i+1] == '>':
			toks = append(toks, dotToken{kind: dotPunct, text: "->", line: line})
			i += 2
		case c == '-' && i+1 < len(src) && src[i+1] == '-':
			toks = append(toks, dotToken{kind: dotPunct, text: "--", line: line})
			i += 2
		case dotIsPunct(c):
			toks = append(toks, dotToken{kind: dotPunct, text: string(c), line: line})
			i++
		default:
			start, startLine := i, line
			for i < len(src) {
				ch := src[i]
				if dotIsSpace(ch) || dotIsPunct(ch) {
					break
				}
				if ch == '/' && i+1 < len(src) && (src[i+1] == '/' || src[i+1] == '*') {
					break
				}
				if ch == '-' && i+1 < len(src) && (src[i+1] == '>' || src[i+1] == '-') {
					break
				}
				i++
			}
			if i == start {
				// Unknown byte: skip it rather than looping forever.
				i++
				continue
			}
			toks = append(toks, dotToken{kind: dotIdent, text: src[start:i], line: startLine})
		}
	}
	toks = append(toks, dotToken{kind: dotEOF, line: line})
	return toks, nil
}

// ---------------------------------------------------------------------------
// Parser
// ---------------------------------------------------------------------------

type dotNode struct {
	path  string // D2 path, e.g. `foo.a`
	shape string // last emitted D2 shape
	label string // last emitted label
}

type dotParser struct {
	toks     []dotToken
	pos      int
	out      []string
	indent   int
	scope    string // D2 path of the enclosing container ("" at root)
	nodes    map[string]*dotNode
	defaults map[string]string // node attribute defaults for the current scope
	subgraph int
}

func (p *dotParser) emit(line string) {
	p.out = append(p.out, strings.Repeat("  ", p.indent)+line)
}

func (p *dotParser) peek() dotToken { return p.toks[p.pos] }

func (p *dotParser) peekAt(n int) dotToken {
	if i := p.pos + n; i < len(p.toks) {
		return p.toks[i]
	}
	return p.toks[len(p.toks)-1]
}

func (p *dotParser) next() dotToken {
	t := p.peek()
	if p.pos < len(p.toks)-1 {
		p.pos++
	}
	return t
}

func (p *dotParser) atPunct(s string) bool {
	t := p.peek()
	return t.kind == dotPunct && t.text == s
}

func (p *dotParser) atPunctAt(n int, s string) bool {
	t := p.peekAt(n)
	return t.kind == dotPunct && t.text == s
}

func (p *dotParser) atKeyword(s string) bool {
	t := p.peek()
	return t.kind == dotIdent && strings.EqualFold(t.text, s)
}

// parse consumes an optional `strict` prefix, an optional graph/digraph header
// and the graph body (braced or bare).
func (p *dotParser) parse() error {
	for p.atKeyword("strict") {
		p.next()
	}
	if p.atKeyword("digraph") || p.atKeyword("graph") {
		p.next()
		if t := p.peek(); t.kind == dotIdent || t.kind == dotString {
			p.next()
		}
	}
	if p.atPunct("{") {
		p.next()
		if err := p.parseBody(); err != nil {
			return err
		}
		if !p.atPunct("}") {
			return fmt.Errorf("dot: missing closing brace")
		}
		p.next()
		return nil
	}
	return p.parseBody()
}

func (p *dotParser) parseBody() error {
	for {
		t := p.peek()
		switch {
		case t.kind == dotEOF:
			return nil
		case t.kind == dotPunct && t.text == "}":
			return nil
		case t.kind == dotPunct && (t.text == ";" || t.text == ","):
			p.next()
		case t.kind == dotPunct && t.text == "{":
			p.next()
			if err := p.parseBlock(""); err != nil {
				return err
			}
		case t.kind == dotIdent && strings.EqualFold(t.text, "subgraph"):
			if err := p.parseSubgraph(); err != nil {
				return err
			}
		case t.kind == dotIdent && strings.EqualFold(t.text, "node") && p.atPunctAt(1, "["):
			p.next()
			attrs, err := p.parseAttrLists()
			if err != nil {
				return err
			}
			p.mergeDefaults(attrs)
		case t.kind == dotIdent && strings.EqualFold(t.text, "graph") && p.atPunctAt(1, "["):
			p.next()
			attrs, err := p.parseAttrLists()
			if err != nil {
				return err
			}
			p.applyGraphAttrs(attrs)
		case t.kind == dotIdent && strings.EqualFold(t.text, "edge") && p.atPunctAt(1, "["):
			// Edge defaults have no D2 equivalent; consume and drop them.
			p.next()
			if _, err := p.parseAttrLists(); err != nil {
				return err
			}
		case t.kind == dotPunct:
			return fmt.Errorf("dot: unexpected %q at line %d", t.text, t.line)
		default:
			if err := p.parseStatement(); err != nil {
				return err
			}
		}
	}
}

func (p *dotParser) mergeDefaults(attrs map[string]string) {
	if p.defaults == nil {
		p.defaults = make(map[string]string)
	}
	for k, v := range attrs {
		p.defaults[k] = v
	}
}

func (p *dotParser) seedDefaults() map[string]string {
	if p.defaults == nil {
		return nil
	}
	clone := make(map[string]string, len(p.defaults))
	for k, v := range p.defaults {
		clone[k] = v
	}
	return clone
}

func (p *dotParser) parseSubgraph() error {
	p.next() // subgraph
	name := ""
	if t := p.peek(); t.kind == dotString || (t.kind == dotIdent && !strings.EqualFold(t.text, "subgraph")) {
		name = t.text
		p.next()
	}
	if !p.atPunct("{") {
		if name != "" {
			// A reference to a previously declared subgraph: nothing to emit.
			return nil
		}
		return fmt.Errorf("dot: expected '{' after subgraph at line %d", p.peek().line)
	}
	p.next()
	return p.parseBlock(name)
}

// parseBlock parses a braced body (the opening brace is already consumed) and
// emits it as a D2 container.
func (p *dotParser) parseBlock(name string) error {
	base := name
	if strings.HasPrefix(base, "cluster_") {
		base = strings.TrimPrefix(base, "cluster_")
	}
	if base == "" {
		p.subgraph++
		base = fmt.Sprintf("subgraph%d", p.subgraph)
	}
	seg := dotID(base)
	path := seg
	if p.scope != "" {
		path = p.scope + "." + seg
	}
	p.emit(seg + ": {")
	p.indent++
	oldScope, oldDefaults := p.scope, p.defaults
	p.scope, p.defaults = path, p.seedDefaults()
	err := p.parseBody()
	p.scope, p.defaults = oldScope, oldDefaults
	p.indent--
	p.emit("}")
	if err != nil {
		return err
	}
	if !p.atPunct("}") {
		return fmt.Errorf("dot: missing closing brace for subgraph %q", base)
	}
	p.next()
	return nil
}

func (p *dotParser) parseStatement() error {
	id, err := p.parseID()
	if err != nil {
		return err
	}
	if err := p.skipPorts(); err != nil {
		return err
	}
	if p.atPunct("=") {
		p.next()
		val, err := p.parseID()
		if err != nil {
			return err
		}
		p.applyGraphAttr(id, val)
		return nil
	}
	if p.atPunct("[") {
		attrs, err := p.parseAttrLists()
		if err != nil {
			return err
		}
		p.emitNode(id, attrs)
		if p.atPunct("->") || p.atPunct("--") {
			return p.parseEdgeChain(id)
		}
		return nil
	}
	if p.atPunct("->") || p.atPunct("--") {
		return p.parseEdgeChain(id)
	}
	p.emitNode(id, nil)
	return nil
}

// skipPorts consumes optional `:port` (and `:compass`) suffixes on a node id.
func (p *dotParser) skipPorts() error {
	for p.atPunct(":") {
		p.next()
		if _, err := p.parseID(); err != nil {
			return err
		}
	}
	return nil
}

func (p *dotParser) parseID() (string, error) {
	t := p.peek()
	if t.kind != dotIdent && t.kind != dotString {
		return "", fmt.Errorf("dot: expected identifier at line %d, got %q", t.line, t.text)
	}
	p.next()
	if t.text == "" {
		return "", fmt.Errorf("dot: empty identifier at line %d", t.line)
	}
	return t.text, nil
}

func (p *dotParser) parseAttrLists() (map[string]string, error) {
	attrs := make(map[string]string)
	for p.atPunct("[") {
		p.next()
		for {
			t := p.peek()
			switch {
			case t.kind == dotEOF:
				return nil, fmt.Errorf("dot: unterminated attribute list")
			case t.kind == dotPunct && t.text == "]":
				p.next()
			case t.kind == dotPunct && (t.text == "," || t.text == ";"):
				p.next()
				continue
			case t.kind == dotPunct:
				return nil, fmt.Errorf("dot: unexpected %q in attribute list at line %d", t.text, t.line)
			default:
				key := strings.ToLower(t.text)
				p.next()
				val := "true"
				if p.atPunct("=") {
					p.next()
					v, err := p.parseID()
					if err != nil {
						return nil, err
					}
					val = v
				}
				attrs[key] = val
				continue
			}
			break
		}
	}
	return attrs, nil
}

func (p *dotParser) parseEdgeChain(first string) error {
	ids := []string{first}
	label := ""
	for p.atPunct("->") || p.atPunct("--") {
		p.next()
		id, err := p.parseID()
		if err != nil {
			return err
		}
		if err := p.skipPorts(); err != nil {
			return err
		}
		ids = append(ids, id)
		if p.atPunct("[") {
			attrs, err := p.parseAttrLists()
			if err != nil {
				return err
			}
			if label == "" {
				label = attrs["label"]
			}
		}
	}
	for i := 0; i+1 < len(ids); i++ {
		src := p.reference(ids[i])
		dst := p.reference(ids[i+1])
		if label != "" {
			p.emit(src + " -> " + dst + ": " + dotLabel(label))
		} else {
			p.emit(src + " -> " + dst)
		}
	}
	return nil
}

func (p *dotParser) applyGraphAttrs(attrs map[string]string) {
	if v, ok := attrs["rankdir"]; ok {
		p.applyGraphAttr("rankdir", v)
	}
	if v, ok := attrs["label"]; ok {
		p.applyGraphAttr("label", v)
	}
}

func (p *dotParser) applyGraphAttr(name, value string) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "rankdir":
		if dir, ok := dotDirections[strings.ToUpper(strings.TrimSpace(value))]; ok {
			p.emit("direction: " + dir)
		}
	case "label":
		if p.scope == "" {
			p.emit("title: " + dotLabel(value))
		} else {
			// Inside a container this resolves to `<container>.label`.
			p.emit("label: " + dotLabel(value))
		}
	}
}

// ---------------------------------------------------------------------------
// D2 emission
// ---------------------------------------------------------------------------

func (p *dotParser) nodeState(id string) (*dotNode, bool) {
	if n, ok := p.nodes[id]; ok {
		return n, false
	}
	seg := dotID(id)
	path := seg
	if p.scope != "" {
		path = p.scope + "." + seg
	}
	n := &dotNode{path: path}
	p.nodes[id] = n
	return n, true
}

func (p *dotParser) rel(path string) string {
	if p.scope == "" {
		return path
	}
	prefix := p.scope + "."
	if strings.HasPrefix(path, prefix) {
		return strings.TrimPrefix(path, prefix)
	}
	return path
}

func (p *dotParser) setShape(n *dotNode, shape string) {
	mapped := dotShape(shape)
	if n.shape == mapped {
		return
	}
	n.shape = mapped
	p.emit(p.rel(n.path) + ".shape: " + mapped)
}

func (p *dotParser) setLabel(n *dotNode, label string) {
	if n.label == label {
		return
	}
	n.label = label
	p.emit(p.rel(n.path) + ": " + dotLabel(label))
}

// emitNode emits an explicit node statement together with its attributes.
func (p *dotParser) emitNode(id string, attrs map[string]string) {
	n, isNew := p.nodeState(id)
	if isNew {
		p.emit(p.rel(n.path))
		if _, ok := attrs["shape"]; !ok {
			if v, ok := p.defaults["shape"]; ok {
				p.setShape(n, v)
			}
		}
		if _, ok := attrs["label"]; !ok {
			if v, ok := p.defaults["label"]; ok {
				p.setLabel(n, v)
			}
		}
	}
	if v, ok := attrs["shape"]; ok {
		p.setShape(n, v)
	}
	if v, ok := attrs["label"]; ok {
		p.setLabel(n, v)
	}
}

// reference registers a node used as an edge endpoint and returns the D2 id to
// reference it by. Node defaults are applied the first time the node is seen.
func (p *dotParser) reference(id string) string {
	n, isNew := p.nodeState(id)
	if isNew {
		if v, ok := p.defaults["shape"]; ok {
			p.setShape(n, v)
		}
		if v, ok := p.defaults["label"]; ok {
			p.setLabel(n, v)
		}
	}
	return p.rel(n.path)
}

// ---------------------------------------------------------------------------
// DOT -> D2 value mapping
// ---------------------------------------------------------------------------

var dotDirections = map[string]string{
	"LR": "right",
	"RL": "left",
	"TB": "down",
	"BT": "up",
}

var dotShapes = map[string]string{
	"box":            "rectangle",
	"rect":           "rectangle",
	"rectangle":      "rectangle",
	"ellipse":        "oval",
	"oval":           "oval",
	"circle":         "circle",
	"doublecircle":   "circle",
	"point":          "circle",
	"diamond":        "diamond",
	"hexagon":        "hexagon",
	"cylinder":       "cylinder",
	"cylinder_shape": "cylinder",
	"box3d":          "cylinder",
	"parallelogram":  "parallelogram",
	"trapezium":      "trapezoid",
	"house":          "rectangle",
	"invhouse":       "rectangle",
	"record":         "rectangle",
	"mrecord":        "rectangle",
	"plaintext":      "rectangle",
	"none":           "rectangle",
	"cloud":          "cloud",
	"folder":         "rectangle",
	"note":           "rectangle",
	"tab":            "rectangle",
}

// dotShape maps a DOT shape name onto a D2 shape. Unknown shapes fall back to
// rectangle.
func dotShape(shape string) string {
	if s, ok := dotShapes[strings.ToLower(strings.TrimSpace(shape))]; ok {
		return s
	}
	return "rectangle"
}

// dotID quotes an identifier when it contains anything outside [A-Za-z0-9_].
func dotID(id string) string {
	if id == "" {
		return `""`
	}
	for _, r := range id {
		if r == '_' ||
			(r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') {
			continue
		}
		return dotQuote(id)
	}
	return id
}

func dotQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
		case '\t':
			b.WriteByte(' ')
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// dotLabel renders a label value, quoting it only when raw text would be
// ambiguous in D2.
func dotLabel(s string) string {
	if s == "" {
		return `""`
	}
	if strings.TrimSpace(s) != s || strings.ContainsAny(s, "\"{}[];|#\\\n\r\t") {
		return dotQuote(s)
	}
	return s
}
