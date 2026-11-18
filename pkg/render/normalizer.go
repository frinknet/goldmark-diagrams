package render

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/goldmark-diagrams/goldmark-diagrams/pkg/cache"
	"github.com/goldmark-diagrams/goldmark-diagrams/pkg/engines"
)

// diagramSeriesCount is the number of --diagram-series-N slots the palette publishes.
const diagramSeriesCount = 8

// textContentRE extracts the text of a <text> element for <title> tooltips.
var textContentRE = regexp.MustCompile(`(?s)<text\b[^>]*>(.*?)</text>`)

// innerTagRE strips nested markup from a <text> body.
var innerTagRE = regexp.MustCompile(`(?s)<[^>]*>`)

// NormalizeSVG rewrites a raw engine SVG into the single <svg> output contract:
// a responsive viewBox, no hardcoded pixel size, the svg-diagram-svg class, and
// data-* hooks on the interactive elements. It never emits <canvas>, <img>, or
// inline <script>.
func NormalizeSVG(rawSVG string, meta engines.Meta) (string, error) {
	start := strings.Index(rawSVG, "<svg")
	if start < 0 {
		return "", fmt.Errorf("no <svg> element in %s output", meta.Engine)
	}
	svg := rawSVG[start:]
	end := tagEndFrom(svg, 0)
	if end < 0 {
		return "", fmt.Errorf("unterminated <svg> element in %s output", meta.Engine)
	}
	name, attrs := parseOpenTag(svg[:end])
	body := svg[end+1:]

	// A nested <svg> closes the document: nothing follows the root element.
	if close := strings.LastIndex(body, "</svg>"); close >= 0 {
		body = body[:close]
	}

	attrs = ensureViewBox(attrs, meta)
	attrs = removeAttrs(attrs, "width", "height")
	attrs = setClass(attrs, "svg-diagram-svg")
	attrs = setAttr(attrs, "role", "img")
	if attrValue(attrs, "aria-label") == "" {
		attrs = setAttr(attrs, "aria-label", meta.Lang+" diagram")
	}
	if attrValue(attrs, "xmlns") == "" {
		attrs = setAttr(attrs, "xmlns", "http://www.w3.org/2000/svg")
	}
	attrs = setAttr(attrs, "style", appendStyle(attrValue(attrs, "style"), engineVarAliases(meta.Engine)))

	body = stampInteractive(body, meta)
	switch meta.Engine {
	case engines.EnginePele:
		body = insertTitles(body)
	case engines.EngineD2:
		body = nameD2Items(body)
		body = unifyD2Palette(body)
	}
	out := renderOpenTag(name, attrs) + ">" + body
	return out + "</svg>", nil
}

type svgAttr struct {
	Name  string
	Value string
	Bare  bool // present with no value
}

// parseOpenTag splits "<svg a=\"1\" b>" into the element name and its attributes.
func parseOpenTag(tag string) (string, []svgAttr) {
	s := strings.TrimSpace(tag)
	s = strings.TrimPrefix(s, "<")
	s = strings.TrimSuffix(strings.TrimSpace(s), "/")
	i := 0
	for i < len(s) && !isSpaceByte(s[i]) {
		i++
	}
	return s[:i], parseAttrs(s[i:])
}

func parseAttrs(s string) []svgAttr {
	var attrs []svgAttr
	i := 0
	for i < len(s) {
		for i < len(s) && isSpaceByte(s[i]) {
			i++
		}
		if i >= len(s) {
			break
		}
		start := i
		for i < len(s) && !isSpaceByte(s[i]) && s[i] != '=' {
			i++
		}
		name := s[start:i]
		if name == "" {
			i++
			continue
		}
		for i < len(s) && isSpaceByte(s[i]) {
			i++
		}
		if i >= len(s) || s[i] != '=' {
			attrs = append(attrs, svgAttr{Name: name, Bare: true})
			continue
		}
		i++
		for i < len(s) && isSpaceByte(s[i]) {
			i++
		}
		if i < len(s) && (s[i] == '"' || s[i] == '\'') {
			quote := s[i]
			i++
			vs := i
			for i < len(s) && s[i] != quote {
				i++
			}
			attrs = append(attrs, svgAttr{Name: name, Value: s[vs:i]})
			if i < len(s) {
				i++
			}
			continue
		}
		vs := i
		for i < len(s) && !isSpaceByte(s[i]) {
			i++
		}
		attrs = append(attrs, svgAttr{Name: name, Value: s[vs:i]})
	}
	return attrs
}

func renderOpenTag(name string, attrs []svgAttr) string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(name)
	for _, a := range attrs {
		b.WriteByte(' ')
		b.WriteString(a.Name)
		if a.Bare {
			continue
		}
		b.WriteString(`="`)
		b.WriteString(a.Value)
		b.WriteByte('"')
	}
	return b.String()
}

// tagEndFrom returns the index of the '>' closing the tag at lt, ignoring '>'
// inside quoted attribute values. It returns -1 when the tag is unterminated.
func tagEndFrom(s string, lt int) int {
	var quote byte
	for i := lt + 1; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			return i
		}
	}
	return -1
}

func ensureViewBox(attrs []svgAttr, meta engines.Meta) []svgAttr {
	if attrValue(attrs, "viewBox") != "" {
		return attrs
	}
	w, h := meta.Width, meta.Height
	if v := attrValue(attrs, "width"); v != "" {
		if f, err := parseLength(v); err == nil && f > 0 {
			w = f
		}
	}
	if v := attrValue(attrs, "height"); v != "" {
		if f, err := parseLength(v); err == nil && f > 0 {
			h = f
		}
	}
	if w <= 0 {
		w = 800
	}
	if h <= 0 {
		h = 480
	}
	return setAttr(attrs, "viewBox", fmt.Sprintf("0 0 %s %s", trimFloat(w), trimFloat(h)))
}

func removeAttrs(attrs []svgAttr, names ...string) []svgAttr {
	drop := make(map[string]bool, len(names))
	for _, n := range names {
		drop[n] = true
	}
	out := attrs[:0]
	for _, a := range attrs {
		if drop[a.Name] {
			continue
		}
		out = append(out, a)
	}
	return out
}

func attrValue(attrs []svgAttr, name string) string {
	for _, a := range attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

func setAttr(attrs []svgAttr, name, value string) []svgAttr {
	for i, a := range attrs {
		if a.Name == name {
			attrs[i] = svgAttr{Name: name, Value: value}
			return attrs
		}
	}
	return append(attrs, svgAttr{Name: name, Value: value})
}

func setClass(attrs []svgAttr, token string) []svgAttr {
	classes := classTokens(attrs)
	for _, c := range classes {
		if c == token {
			return attrs
		}
	}
	classes = append(classes, token)
	return setAttr(attrs, "class", strings.Join(classes, " "))
}

func classTokens(attrs []svgAttr) []string {
	return strings.Fields(attrValue(attrs, "class"))
}

func hasClass(classes []string, want string) bool {
	for _, c := range classes {
		if c == want {
			return true
		}
	}
	return false
}

func appendStyle(existing, extra string) string {
	if existing == "" {
		return extra
	}
	if strings.HasSuffix(existing, ";") {
		return existing + extra
	}
	return existing + ";" + extra
}

// engineVarAliases declares the bridge variables an engine needs on the root
// <svg> so every diagram type answers to the same --diagram-* palette. The --diagram-*
// defaults themselves live in the reference stylesheet (DefaultCSS), so an
// author can retheme every engine from one place while the engine's own default
// remains the fallback when that stylesheet is not loaded.
func engineVarAliases(engine string) string {
	if engine != engines.EnginePele {
		return ""
	}
	// pele publishes --pele-*; point them at the shared vocabulary. pele uses
	// --pele-bg as its node/card fill, so it maps onto a surface rather than the
	// transparent canvas colour.
	pairs := []struct{ pele, shared, fallback string }{
		{"--pele-bg", "--diagram-surface", "#fff"},
		{"--pele-fg", "--diagram-fg", "#1f1f1f"},
		{"--pele-muted", "--diagram-muted", "#6e6e6e"},
		{"--pele-line", "--diagram-line", "#1f1f1f"},
		{"--pele-surface", "--diagram-surface-alt", "#f3f3f3"},
		{"--pele-surface-alt", "--diagram-surface-alt", "#e6e6e6"},
		{"--pele-border", "--diagram-border", "#8a8a8a"},
		{"--pele-accent", "--diagram-accent", "#5b7bd5"},
	}
	var b strings.Builder
	for i, p := range pairs {
		if i > 0 {
			b.WriteByte(';')
		}
		fmt.Fprintf(&b, "%s:var(%s,%s)", p.pele, p.shared, p.fallback)
	}
	return b.String()
}

// stampInteractive walks the SVG body and adds the predictable data-* hooks the
// reference CSS and the server feedback loop rely on. Every diagram type ends up
// with the same vocabulary: data-item on each primary mark, data-series on
// multi-series marks, and data-column-id/data-card-id for kanban.
func stampInteractive(body string, meta engines.Meta) string {
	var out strings.Builder
	state := &hookState{
		seriesByColor: map[string]int{},
		seriesSeen:    map[int]int{},
	}
	// stack[i] records whether the element at that depth is a defs-only
	// container (marker, mask, ...) or already carries a data-item hook, so
	// nested drawing primitives are not stamped a second time.
	var stack []hookFrame
	i := 0
	for i < len(body) {
		lt := strings.IndexByte(body[i:], '<')
		if lt < 0 {
			out.WriteString(body[i:])
			break
		}
		lt += i
		out.WriteString(body[i:lt])

		if lt+1 < len(body) && (body[lt+1] == '/' || body[lt+1] == '?') {
			end := tagEndFrom(body, lt)
			if end < 0 {
				out.WriteString(body[lt:])
				break
			}
			out.WriteString(body[lt : end+1])
			if body[lt+1] == '/' && len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			i = end + 1
			continue
		}
		if strings.HasPrefix(body[lt:], "<![CDATA[") {
			end := strings.Index(body[lt:], "]]>")
			if end < 0 {
				out.WriteString(body[lt:])
				break
			}
			out.WriteString(body[lt : lt+end+3])
			i = lt + end + 3
			continue
		}
		if strings.HasPrefix(body[lt:], "<!--") {
			end := strings.Index(body[lt:], "-->")
			if end < 0 {
				out.WriteString(body[lt:])
				break
			}
			out.WriteString(body[lt : lt+end+3])
			i = lt + end + 3
			continue
		}

		end := tagEndFrom(body, lt)
		if end < 0 {
			out.WriteString(body[lt:])
			break
		}
		name, attrs := parseOpenTag(body[lt:end])
		if name == "style" || name == "script" {
			// Copy element content verbatim: it is CSS/JS, not markup to stamp.
			closeTag := "</" + name + ">"
			rel := strings.Index(body[end+1:], closeTag)
			if rel < 0 {
				out.WriteString(body[lt:])
				break
			}
			stop := end + 1 + rel + len(closeTag)
			out.WriteString(body[lt:stop])
			i = stop
			continue
		}

		nested := len(stack) > 0 && stack[len(stack)-1].closed
		inDefs := len(stack) > 0 && stack[len(stack)-1].defs
		owner := ""
		if len(stack) > 0 {
			owner = stack[len(stack)-1].owner
		}
		if meta.Engine == engines.EngineD2 {
			if id := d2OwnerID(name, attrs); id != "" {
				owner = id
			}
		}
		injected := dataHooks(name, attrs, meta, state, nested, inDefs, owner)
		for _, a := range injected {
			switch a.Name {
			case "data-role":
				attrs = setClass(attrs, "svg-"+a.Value)
			case "data-series":
				attrs = setClass(attrs, "svg-series-"+a.Value)
			}
		}

		out.WriteString(renderOpenTag(name, attrs))
		for _, a := range injected {
			out.WriteByte(' ')
			out.WriteString(a.Name)
			fmt.Fprintf(&out, `="%s"`, html.EscapeString(a.Value))
		}
		out.WriteByte('>')

		if !strings.HasSuffix(strings.TrimSpace(body[lt:end]), "/") {
			stack = append(stack, hookFrame{
				closed: nested || len(injected) > 0,
				defs:   inDefs || isDefinitionContainer(name),
				owner:  owner,
			})
		}
		i = end + 1
	}
	return out.String()
}

// hookFrame tracks whether stamped hooks or definitions enclose the next element.
type hookFrame struct {
	closed bool   // an ancestor is a stamped item, so children are its drawings
	defs   bool   // inside <defs>, <marker>, <mask>, ... (never drawn directly)
	owner  string // the D2 id of the element whose group encloses this one
}

// hookState carries the counters a single stamping pass needs.
type hookState struct {
	items         int
	seriesByColor map[string]int
	seriesSeen    map[int]int
}

func isDefinitionContainer(name string) bool {
	switch name {
	case "defs", "marker", "mask", "clippath", "symbol", "pattern", "filter", "lineargradient", "radialgradient":
		return true
	}
	return false
}

// dataHooks decides which data-* attributes an element earns. owner is the D2
// id of the enclosing element group, which names connections after their
// endpoints; every other engine leaves it empty.
func dataHooks(name string, attrs []svgAttr, meta engines.Meta, state *hookState, nested, inDefs bool, owner string) []svgAttr {
	if inDefs {
		return nil
	}
	classes := classTokens(attrs)
	dataID := attrValue(attrs, "data-id")
	var hooks []svgAttr

	switch {
	case hasClass(classes, "pele-column"):
		// Columns carry only the column hook; cards inside them carry the card
		// hook, and the reference CSS dims siblings, so neither may also be a
		// data-item or hovering a card would dim its own column.
		if slug(dataID) != "" {
			hooks = append(hooks, svgAttr{Name: "data-column-id", Value: slug(dataID)})
			hooks = append(hooks, svgAttr{Name: "data-role", Value: "column"})
		}
	case hasClass(classes, "pele-card"):
		if dataID != "" {
			hooks = append(hooks, svgAttr{Name: "data-card-id", Value: dataID})
			hooks = append(hooks, svgAttr{Name: "data-role", Value: "card"})
		}
	default:
		switch {
		case peleItemRole(classes, dataID) != "":
			hooks = append(hooks,
				svgAttr{Name: "data-item", Value: dataID},
				svgAttr{Name: "data-role", Value: peleItemRole(classes, dataID)})
		case nested:
			// A drawing primitive inside an already-stamped item is part of that
			// item, not an item of its own: D2 puts the shape's rect and path
			// inside the group that already carries the hook.
			return nil
		case hasClass(classes, "shape"):
			// A shape is renamed to its label after stamping; until then its D2
			// id keeps the hook a real identity.
			hooks = append(hooks,
				svgAttr{Name: "data-item", Value: orIndex(slug(owner), &state.items)},
				svgAttr{Name: "data-role", Value: "shape"})
		case hasClass(classes, "connection"):
			conn, _ := d2ConnectionItem(owner)
			hooks = append(hooks,
				svgAttr{Name: "data-item", Value: orIndex(conn, &state.items)},
				svgAttr{Name: "data-role", Value: "connection"})
		case meta.Engine == engines.EngineCharts && isChartMark(name, attrs):
			hooks = append(hooks, chartHooks(attrs, meta, state)...)
		}
	}
	return hooks
}

// chartHooks names a plotted mark with the category label and series slot.
func chartHooks(attrs []svgAttr, meta engines.Meta, state *hookState) []svgAttr {
	color := fillColor(attrs)
	if color == "" {
		return nil
	}
	slot, ok := state.seriesByColor[color]
	if !ok {
		slot = len(state.seriesByColor)
		state.seriesByColor[color] = slot
	}
	slot %= diagramSeriesCount

	// The engine paints one series at a time, in series order, so the Nth mark
	// of a series is its Nth category: the label the spec gives it, or the
	// 1-based category number the axis prints when the spec names none.
	pos := state.seriesSeen[slot]
	state.seriesSeen[slot] = pos + 1
	item := strconv.Itoa(pos + 1)
	if n := len(meta.Labels); n > 0 {
		item = meta.Labels[pos%n]
	}
	weaveSeriesVar(attrs, slot, color)
	return []svgAttr{
		{Name: "data-item", Value: item},
		{Name: "data-series", Value: strconv.Itoa(slot)},
		{Name: "data-role", Value: "mark"},
	}
}

// peleContainers are grouping elements that wrap items rather than being items.
var peleContainers = map[string]bool{
	"pele-nodes": true, "pele-edges": true, "pele-clusters": true,
	"pele-slices": true, "pele-tasks": true, "pele-verts": true,
	"pele-venn-sets": true, "pele-venn-labels": true, "pele-venn-overlaps": true,
	"pele-edge-labels": true, "pele-slice-labels": true, "pele-point-labels": true,
	"pele-axis-labels": true, "pele-cluster-titles": true, "pele-sections": true,
	"pele-grids": true, "pele-tracks": true, "pele-players": true, "pele-lines": true,
	"pele-legends": true, "pele-stems": true,
}

// peleItemRole returns the shared role name for a pele item, or "" when the
// element is not one. pele marks every drawable with a pele-* class and, for the
// primary elements, a data-id; the shortest class remainder wins, which peels
// modifiers such as pele-shape-rect off a pele-node. An element with no data-id
// is skipped rather than numbered, so every hook is a real identity. Labels are
// skipped too: they repeat the id of the item they annotate.
func peleItemRole(classes []string, dataID string) string {
	if dataID == "" {
		return ""
	}
	best := ""
	for _, c := range classes {
		if !strings.HasPrefix(c, "pele-") || peleContainers[c] {
			continue
		}
		rest := strings.TrimPrefix(c, "pele-")
		if rest == "" || strings.Contains(rest, "label") {
			continue
		}
		if best == "" || len(rest) < len(best) {
			best = rest
		}
	}
	return best
}

// isChartMark reports whether an element is a plotted data mark rather than
// chart furniture. Only path and rect carry series data in the charts engine's
// output; text, circles and lines are labels, legend swatches and grid lines.
func isChartMark(name string, attrs []svgAttr) bool {
	if name != "path" && name != "rect" {
		return false
	}
	return fillColor(attrs) != ""
}

// weaveSeriesVar rewrites an inline chart fill so the --diagram-series-N custom
// property governs it, with the engine's own colour as the fallback.
func weaveSeriesVar(attrs []svgAttr, slot int, color string) {
	for i, a := range attrs {
		if a.Name != "style" {
			continue
		}
		needle := "fill:" + color
		if strings.Contains(a.Value, needle) {
			replacement := fmt.Sprintf("fill:var(--diagram-series-%d,%s)", slot, color)
			attrs[i] = svgAttr{Name: "style", Value: strings.Replace(a.Value, needle, replacement, 1)}
		}
		return
	}
}

func isSpaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f'
}

// d2RuleRE matches the palette rules inside D2's embedded stylesheet.
var d2RuleRE = regexp.MustCompile(`\.(fill|stroke)-([A-Za-z0-9]+)\{([^}]*)\}`)

// colorDeclRE matches a paint declaration inside one of those rules.
var colorDeclRE = regexp.MustCompile(`(fill|stroke):(#[0-9A-Fa-f]{3,8})`)

// unifyD2Palette routes D2's theme colours through the shared --diagram-* vocabulary.
// D2 resolves shape paint through class rules in its own <style> block, so the
// rules are rewritten by the palette role their class encodes (its N ramp is
// neutral, B/AA/AB carry the accent and its tints), with each rule's own colour
// kept as the fallback. The same mapping applies to the light and dark blocks, so
// one stylesheet can theme both.
func unifyD2Palette(body string) string {
	var out strings.Builder
	i := 0
	for {
		start := strings.Index(body[i:], "<style")
		if start < 0 {
			break
		}
		start += i
		stop := strings.Index(body[start:], "</style>")
		if stop < 0 {
			break
		}
		stop += start
		out.WriteString(body[i:start])
		out.WriteString(d2RuleRE.ReplaceAllStringFunc(body[start:stop], func(rule string) string {
			m := d2RuleRE.FindStringSubmatch(rule)
			prop, suffix := m[1], m[2]
			decls := colorDeclRE.ReplaceAllStringFunc(m[3], func(decl string) string {
				d := colorDeclRE.FindStringSubmatch(decl)
				return d[1] + ":" + diagramColorVar(d[1], suffix, d[2])
			})
			return "." + prop + "-" + suffix + "{" + decls + "}"
		}))
		i = stop
	}
	out.WriteString(body[i:])
	return out.String()
}

// diagramColorVar maps one D2 palette colour onto a --diagram-* variable. D2's classes
// encode the role: N is the neutral ramp (1 darkest, 7 lightest), and B/AA/AB
// carry the accent at 1-2 and its tints at 3-6.
func diagramColorVar(prop, suffix, hex string) string {
	family, digit := splitPaletteClass(suffix)
	switch family {
	case "N":
		if prop == "stroke" {
			switch digit {
			case "1", "3", "4", "5", "6", "7":
				return "var(--diagram-border, " + hex + ")"
			case "2":
				return "var(--diagram-muted, " + hex + ")"
			}
			return "var(--diagram-border, " + hex + ")"
		}
		switch digit {
		case "1":
			return "var(--diagram-fg, " + hex + ")"
		case "2":
			return "var(--diagram-muted, " + hex + ")"
		case "5":
			return "var(--diagram-surface-alt, " + hex + ")"
		case "6":
			return "var(--diagram-surface, " + hex + ")"
		case "7":
			return "var(--diagram-bg, " + hex + ")"
		}
		return "var(--diagram-border, " + hex + ")"
	}
	switch digit {
	case "1", "2":
		return "var(--diagram-accent, " + hex + ")"
	case "4":
		return "var(--diagram-surface-alt, " + hex + ")"
	case "5", "6":
		return "var(--diagram-surface, " + hex + ")"
	}
	return "var(--diagram-border, " + hex + ")"
}

// splitPaletteClass splits a D2 palette class such as N7, B3 or AA4 into its
// family letter and trailing digit.
func splitPaletteClass(suffix string) (family, digit string) {
	i := len(suffix)
	for i > 0 && suffix[i-1] >= '0' && suffix[i-1] <= '9' {
		i--
	}
	return suffix[:i], suffix[i:]
}

// d2OwnerID decodes the identifier D2 stamps on the group that wraps every shape
// and connection: the base64 of its HTML-escaped id. The group carries no
// readable name, so this is the only way back to "alpha" or "(alpha -> beta)[0]".
func d2OwnerID(name string, attrs []svgAttr) string {
	if name != "g" {
		return ""
	}
	classes := classTokens(attrs)
	if len(classes) == 0 {
		return ""
	}
	raw, err := base64.URLEncoding.DecodeString(classes[0])
	if err != nil {
		return ""
	}
	return html.UnescapeString(string(raw))
}

// d2ConnectionItem turns D2's connection id, "(src -> dst)[n]", into the
// endpoint identity a client-side handler can store: "src->dst", with the index
// kept only for parallel connections between the same pair.
func d2ConnectionItem(id string) (string, bool) {
	if !strings.HasPrefix(id, "(") || !strings.HasSuffix(id, "]") {
		return "", false
	}
	open := strings.LastIndex(id, ")[")
	if open < 0 {
		return "", false
	}
	src, dst, ok := strings.Cut(id[1:open], " -> ")
	if !ok {
		return "", false
	}
	src, dst = strings.TrimSpace(src), strings.TrimSpace(dst)
	if src == "" || dst == "" {
		return "", false
	}
	item := src + "->" + dst
	if n := id[open+2 : len(id)-1]; n != "" && n != "0" {
		item += "[" + n + "]"
	}
	return item, true
}

// nameD2Items replaces a shape's provisional D2 id with the shape's own label,
// which is what a client-side handler needs to survive an edit, and adds a
// <title> child for zero-JS hover. D2 nests the label either inside the shape
// group (tables) or next to it in the wrapper group (flowcharts), so both are
// checked.
func nameD2Items(body string) string {
	type edit struct {
		start, end int
		text       string
	}
	var edits []edit

	i := 0
	for i < len(body) {
		lt := strings.Index(body[i:], "<g ")
		if lt < 0 {
			break
		}
		lt += i
		end := tagEndFrom(body, lt)
		if end < 0 {
			break
		}
		if strings.Contains(body[lt:end+1], `data-role="shape"`) {
			i = end + 1
			continue
		}
		close, ok := matchingGroupEnd(body, end+1)
		if !ok {
			i = end + 1
			continue
		}
		inner := body[end+1 : close]
		shapeRel := strings.Index(inner, "<g ")
		if shapeRel < 0 {
			i = close
			continue
		}
		shapeAbs := end + 1 + shapeRel
		shapeEnd := tagEndFrom(body, shapeAbs)
		shapeTag := body[shapeAbs : shapeEnd+1]
		if shapeEnd < 0 || shapeEnd > close ||
			!strings.Contains(shapeTag, `data-role="shape"`) ||
			!strings.Contains(shapeTag, "data-item=") {
			i = close
			continue
		}

		label := ""
		if shapeClose, ok := matchingGroupEnd(body, shapeEnd+1); ok && shapeClose <= close {
			label = firstText(body[shapeEnd+1 : shapeClose])
		}
		if label == "" {
			label = firstText(inner)
		}
		old := attrValue(mustParseAttrs(shapeTag), "data-item")
		if id := slug(label); id != "" && old != "" {
			edits = append(edits, edit{shapeAbs, shapeEnd + 1,
				strings.Replace(shapeTag, `data-item="`+old+`"`, `data-item="`+id+`"`, 1)})
			if !strings.Contains(inner, "<title>") {
				edits = append(edits, edit{shapeEnd + 1, shapeEnd + 1,
					"<title>" + html.EscapeString(label) + "</title>"})
			}
		}
		i = close
	}

	if len(edits) == 0 {
		return body
	}
	var out strings.Builder
	prev := 0
	for _, e := range edits {
		out.WriteString(body[prev:e.start])
		out.WriteString(e.text)
		prev = e.end
	}
	out.WriteString(body[prev:])
	return out.String()
}

// matchingGroupEnd returns the offset of the </g> that closes the group whose
// content starts at from. It returns false when the group never closes.
func matchingGroupEnd(s string, from int) (int, bool) {
	depth := 1
	i := from
	for i < len(s) {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			return 0, false
		}
		lt += i
		end := tagEndFrom(s, lt)
		if end < 0 {
			return 0, false
		}
		switch elementName(s[lt+1 : end]) {
		case "g":
			depth++
		case "/g":
			depth--
			if depth == 0 {
				return lt, true
			}
		}
		i = end + 1
	}
	return 0, false
}

// elementName extracts a tag name from the text between '<' and '>'.
func elementName(tag string) string {
	t := strings.TrimSpace(tag)
	if strings.HasPrefix(t, "/") {
		return "/" + strings.ToLower(firstWord(t[1:]))
	}
	return strings.ToLower(firstWord(t))
}

func firstWord(s string) string {
	for i := 0; i < len(s); i++ {
		if isSpaceByte(s[i]) || s[i] == '/' {
			return s[:i]
		}
	}
	return s
}

func mustParseAttrs(tag string) []svgAttr {
	_, attrs := parseOpenTag(tag)
	return attrs
}

// parseLength reads a leading number from an SVG length such as "640" or "640px".
func parseLength(v string) (float64, error) {
	s := strings.TrimSpace(v)
	end := 0
	for end < len(s) && (s[end] == '+' || s[end] == '-' || s[end] == '.' || (s[end] >= '0' && s[end] <= '9') || s[end] == 'e' || s[end] == 'E') {
		end++
	}
	if end == 0 {
		return 0, fmt.Errorf("invalid length %q", v)
	}
	return strconv.ParseFloat(s[:end], 64)
}

// trimFloat formats a pixel dimension without a trailing ".0".
func trimFloat(f float64) string {
	if f == math.Trunc(f) && math.Abs(f) < 1e9 {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func orIndex(id string, counter *int) string {
	if id != "" {
		return id
	}
	n := *counter
	*counter = n + 1
	return fmt.Sprintf("%d", n)
}

// fillColor returns the fill colour of a shape that carries visible paint,
// which is how the charts engine distinguishes data marks from grid lines.
func fillColor(attrs []svgAttr) string {
	style := attrValue(attrs, "style")
	i := strings.Index(style, "fill:")
	if i < 0 {
		return ""
	}
	v := style[i+len("fill:"):]
	if j := strings.IndexByte(v, ';'); j >= 0 {
		v = v[:j]
	}
	v = strings.TrimSpace(v)
	switch strings.ToLower(v) {
	case "", "none", "white", "#fff", "#ffffff", "transparent":
		return ""
	}
	return v
}

// insertTitles gives stamped pele groups a <title> child so native zero-JS
// hover tooltips work.
func insertTitles(body string) string {
	var out strings.Builder
	i := 0
	for i < len(body) {
		lt := strings.Index(body[i:], "<g ")
		if lt < 0 {
			break
		}
		lt += i
		end := tagEndFrom(body, lt)
		if end < 0 {
			break
		}
		out.WriteString(body[i : end+1])
		tag := body[lt : end+1]
		if !strings.Contains(tag, "<title") && labeledForTooltip(tag) {
			if closeRel := strings.Index(body[end+1:], "</g>"); closeRel >= 0 {
				if label := firstText(body[end+1 : end+1+closeRel]); label != "" {
					out.WriteString("<title>")
					out.WriteString(html.EscapeString(label))
					out.WriteString("</title>")
				}
			}
		}
		i = end + 1
	}
	out.WriteString(body[i:])
	return out.String()
}

// labeledForTooltip reports whether a stamped group is one a reader can hover.
func labeledForTooltip(tag string) bool {
	return strings.Contains(tag, "data-item=") ||
		strings.Contains(tag, "data-card-id=") ||
		strings.Contains(tag, "data-column-id=")
}

func firstText(inner string) string {
	m := textContentRE.FindStringSubmatch(inner)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(innerTagRE.ReplaceAllString(m[1], ""))
}

// figureAttrs are the data-* names buildFigure writes itself, so a fence
// attribute of the same name cannot emit a duplicate attribute.
var figureAttrs = map[string]bool{
	"lang":   true,
	"engine": true,
	"theme":  true,
	"hash":   true,
}

// fenceTheme lets an info-string theme attribute override the configured theme,
// so ` ```mermaid theme=dark ` both selects the engine theme and labels the
// figure.
func fenceTheme(configured string, attrs map[string]string) string {
	if t := strings.TrimSpace(attrs["theme"]); t != "" {
		return t
	}
	return configured
}

// buildFigure wraps a normalized svg in the standardized <figure> element and
// stamps the fence attributes onto it.
func buildFigure(svg string, meta engines.Meta) string {
	var b strings.Builder
	b.WriteString(`<figure class="svg-diagram"`)
	writeAttr(&b, "data-lang", meta.Lang)
	writeAttr(&b, "data-engine", meta.Engine)
	writeAttr(&b, "data-theme", themeOrDefault(meta.Theme))
	writeAttr(&b, "data-hash", cache.ShortHash(svg))
	for _, k := range sortedKeys(meta.Attrs) {
		name := sanitizeAttrName(k)
		if name == "" || figureAttrs[name] {
			// Reserved names are already on the figure; a fence attribute must
			// not emit a second copy of it.
			continue
		}
		writeAttr(&b, "data-"+name, meta.Attrs[k])
	}
	b.WriteString(">")
	b.WriteString(svg)
	b.WriteString("</figure>")
	return b.String()
}

func writeAttr(b *strings.Builder, name, value string) {
	fmt.Fprintf(b, ` %s="%s"`, name, html.EscapeString(value))
}

func themeOrDefault(theme string) string {
	if theme == "" {
		return "auto"
	}
	return theme
}

func sortedKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sanitizeAttrName turns a fence attribute key into a safe data-* suffix.
func sanitizeAttrName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// slug lowercases a label into a css-safe identifier for data-column-id.
func slug(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// Fragment is the shared render path used by the AST transformer and by
// RenderStandalone. It never returns an error unless the configured ErrorMode
// is ErrorFail.
func Fragment(ctx context.Context, cfg *Config, lang, source string, attrs map[string]string, line int) (string, error) {
	canonical := engines.NormalizeLang(lang)
	if canonical == "" {
		canonical = strings.ToLower(strings.TrimSpace(lang))
	}
	attrs = withConfiguredID(attrs, cfg.ID)
	theme := fenceTheme(cfg.Theme, attrs)
	key := cache.Key(canonical, theme, attrsKey(attrs), source)
	if cfg.Cache != nil {
		if cached, ok := cfg.Cache.Get(key); ok {
			return cached, nil
		}
	}

	meta := engines.Meta{
		Lang:     canonical,
		Engine:   engines.EngineForLang(canonical),
		Theme:    theme,
		D2Layout: cfg.D2Layout,
		Attrs:    attrs,
		Line:     line,
	}

	rctx := ctx
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		rctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	res, err := engines.Dispatch(rctx, meta, source)
	if err == nil {
		meta.Engine = res.Engine
		meta.Labels = res.Labels
		meta.Width, meta.Height = res.Width, res.Height
		if res.SVG, err = NormalizeSVG(res.SVG, meta); err != nil {
			res.SVG = ""
		}
	}
	if err != nil {
		switch cfg.ErrorMode {
		case ErrorFail:
			return "", fmt.Errorf("diagrams: %s: %w", canonical, err)
		case ErrorCodeBlock:
			return codeBlockHTML(canonical, source), nil
		default:
			frag := buildFigure(errorSVGCard(canonical, line, err), meta)
			if cfg.Cache != nil {
				cfg.Cache.Set(key, frag)
			}
			return frag, nil
		}
	}

	frag := buildFigure(res.SVG, meta)
	if cfg.Cache != nil {
		cfg.Cache.Set(key, frag)
	}
	return frag, nil
}

// withConfiguredID folds the WithID option into the fence attributes so a
// standalone render carries the data-id the feedback loop needs.
func withConfiguredID(attrs map[string]string, id string) map[string]string {
	if id == "" {
		return attrs
	}
	if _, ok := attrs["id"]; ok {
		return attrs
	}
	merged := make(map[string]string, len(attrs)+1)
	for k, v := range attrs {
		merged[k] = v
	}
	merged["id"] = id
	return merged
}

// attrsKey renders the fence attributes in a stable order for the cache key.
func attrsKey(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}
	var b strings.Builder
	for i, k := range sortedKeys(attrs) {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(attrs[k])
	}
	return b.String()
}
