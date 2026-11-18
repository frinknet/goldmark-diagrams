package transpilers

import (
	"fmt"
	"regexp"
	"strings"
)

// PlantUMLToD2 translates PlantUML source into D2 source.
func PlantUMLToD2(src string) (string, error) {
	switch pumlForm(src) {
	case "gantt", "mindmap":
		// These forms are rendered through Mermaid; see PlantUMLToMermaid.
		return "", fmt.Errorf("plantuml: unsupported diagram")
	}
	p := &pumlD2{
		nodes:    make(map[string]*pumlNode),
		ids:      make(map[string]string),
		taken:    make(map[string]string),
		blockIDs: make(map[string]bool),
	}
	if err := p.run(pumlLines(src)); err != nil {
		return "", err
	}
	p.closeAll()
	if len(p.out) == 0 {
		return "", fmt.Errorf("plantuml: no statements")
	}
	return strings.Join(p.out, "\n") + "\n", nil
}

// PlantUMLToMermaid translates the PlantUML gantt and mindmap diagram forms
// into Mermaid source (pele-go renders Mermaid; D2 cannot draw these).
func PlantUMLToMermaid(src string) (string, error) {
	switch pumlForm(src) {
	case "gantt":
		return pumlGanttToMermaid(pumlLines(src))
	case "mindmap":
		return pumlMindmapToMermaid(pumlLines(src))
	}
	return "", fmt.Errorf("plantuml: unsupported diagram")
}

// ---------------------------------------------------------------------------
// Shared source scrubbing
// ---------------------------------------------------------------------------

var (
	pumlBlockCommentRe = regexp.MustCompile(`(?s)/'.*?'/`)
	pumlAsRe           = regexp.MustCompile(`(?i)^(.*?)\s+as\s+"?([^"\s]+)"?$`)
	pumlBracketRe      = regexp.MustCompile(`^\[(.+)\]$`)

	pumlGanttProjectRe = regexp.MustCompile(`(?i)^project\s+starts\s+(?:on\s+|at\s+)?(\d{4}-\d{2}-\d{2})`)
	pumlGanttSectionRe = regexp.MustCompile(`^--\s*(.*?)\s*--$`)
	pumlGanttTaskRe    = regexp.MustCompile(`^\[([^\]]+)\]\s*(.*)$`)
	pumlGanttLastsRe   = regexp.MustCompile(`(?i)lasts?\s+(?:(\d+)\s*weeks?|(\d+)\s*days?)`)
	pumlGanttStartsRe  = regexp.MustCompile(`(?i)starts?\s+(?:at\s+|on\s+)?(\d{4}-\d{2}-\d{2})`)
	pumlGanttAfterRe   = regexp.MustCompile(`(?i)starts?\s+after\s+\[([^\]]+)\]`)
)

// pumlForm reports which `@start...` marker opens the source, if any.
func pumlForm(src string) string {
	for _, raw := range strings.Split(src, "\n") {
		t := strings.TrimSpace(raw)
		if t == "" || strings.HasPrefix(t, "'") {
			continue
		}
		low := strings.ToLower(t)
		switch {
		case strings.HasPrefix(low, "@startgantt"):
			return "gantt"
		case strings.HasPrefix(low, "@startmindmap"):
			return "mindmap"
		case strings.HasPrefix(low, "@startuml"):
			return "uml"
		case strings.HasPrefix(low, "@start"):
			return "other"
		}
		return "bare"
	}
	return ""
}

// pumlLines splits source into trimmed, comment-free, marker-free lines.
func pumlLines(src string) []string {
	src = pumlBlockCommentRe.ReplaceAllString(src, "")
	raw := strings.Split(src, "\n")
	out := make([]string, 0, len(raw))
	for _, l := range raw {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "'") {
			continue
		}
		low := strings.ToLower(l)
		if strings.HasPrefix(low, "@start") || strings.HasPrefix(low, "@end") {
			continue
		}
		out = append(out, l)
	}
	return out
}

// ---------------------------------------------------------------------------
// PlantUML -> D2
// ---------------------------------------------------------------------------

type pumlNode struct {
	label string
	shape string
}

type pumlD2 struct {
	out      []string
	indent   int
	nodes    map[string]*pumlNode
	ids      map[string]string // raw name -> D2 id
	taken    map[string]string // D2 id -> raw name
	blocks   []string
	blockIDs map[string]bool
	gen      int
}

func (p *pumlD2) emit(line string) {
	p.out = append(p.out, strings.Repeat("  ", p.indent)+line)
}

// shapes for the declaration keywords shared by sequence, component, state and
// usecase diagrams. An empty shape keeps the D2 default (rectangle).
var pumlDeclShapes = map[string]string{
	"participant": "",
	"actor":       "person",
	"boundary":    "",
	"control":     "",
	"entity":      "",
	"collections": "",
	"database":    "cylinder",
	"queue":       "queue",
	"component":   "",
	"state":       "",
	"usecase":     "oval",
	"artifact":    "",
	"cloud":       "cloud",
	"folder":      "",
	"frame":       "",
	"node":        "",
	"package":     "",
	"storage":     "cylinder",
}

var pumlBlockKinds = map[string]bool{
	"alt":      true,
	"opt":      true,
	"loop":     true,
	"group":    true,
	"par":      true,
	"break":    true,
	"critical": true,
	"while":    true,
	"fork":     true,
}

var pumlSkipWords = map[string]bool{
	"skinparam":   true,
	"hide":        true,
	"show":        true,
	"scale":       true,
	"header":      true,
	"footer":      true,
	"legend":      true,
	"endlegend":   true,
	"caption":     true,
	"newpage":     true,
	"together":    true,
	"allowmixing": true,
	"sprite":      true,
	"activate":    true,
	"deactivate":  true,
	"destroy":     true,
	"create":      true,
	"autonumber":  true,
	"return":      true,
	"ref":         true,
	"box":         true,
	"endbox":      true,
	"date":        true,
	"salt":        true,
	"endstyle":    true,
	"endtitle":    true,
}

func (p *pumlD2) run(lines []string) error {
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		low := strings.ToLower(line)
		if low == "title" || strings.HasPrefix(low, "title ") || strings.HasPrefix(low, "title:") {
			i = p.handleTitle(lines, i)
			continue
		}
		if strings.HasPrefix(low, "left to right direction") {
			p.emit("direction: right")
			continue
		}
		if strings.HasPrefix(low, "top to bottom direction") {
			p.emit("direction: down")
			continue
		}
		if strings.HasPrefix(low, "==") {
			continue
		}
		if pumlSkipLine(low) {
			continue
		}
		word := pumlFirstWord(low)
		switch {
		case pumlBlockKinds[word]:
			rest := strings.TrimSpace(line[len(word):])
			p.openBlock(word, rest)
			continue
		case word == "end":
			p.closeBlock()
			continue
		case word == "else" || word == "also":
			rest := strings.TrimSpace(line[len(word):])
			p.closeBlock()
			p.openBlock(word, rest)
			continue
		}
		if p.declare(line) {
			continue
		}
		if next, ok := p.handleNote(lines, i); ok {
			i = next
			continue
		}
		if next, ok := p.handleClass(lines, i); ok {
			i = next
			continue
		}
		if p.handleRelation(line) {
			continue
		}
		if m := pumlBracketRe.FindStringSubmatch(line); m != nil {
			name := cleanPumlName(m[1])
			p.emitNode(p.idFor(name), name, "", true)
			continue
		}
		// Anything else (annotations, layout hints, ...) is ignored.
	}
	return nil
}

func pumlFirstWord(low string) string {
	if i := strings.IndexAny(low, " \t:"); i >= 0 {
		return low[:i]
	}
	return low
}

func pumlSkipLine(low string) bool {
	if strings.HasPrefix(low, "!") || strings.HasPrefix(low, "<") || strings.HasPrefix(low, "|") {
		return true
	}
	return pumlSkipWords[pumlFirstWord(low)]
}

func (p *pumlD2) handleTitle(lines []string, i int) int {
	line := strings.TrimSpace(lines[i])
	rest := strings.TrimSpace(strings.TrimPrefix(line[len("title"):], ":"))
	if rest == "" {
		var parts []string
		j := i + 1
		for ; j < len(lines); j++ {
			t := strings.TrimSpace(lines[j])
			if strings.EqualFold(t, "end title") || strings.EqualFold(t, "endtitle") {
				break
			}
			parts = append(parts, t)
		}
		rest = strings.Join(parts, " ")
		i = j - 1
	}
	if rest != "" {
		p.emit("title: " + pumlLabel(rest))
	}
	return i
}

// declare handles the keyword declarations of sequence, class, component,
// state and usecase diagrams.
func (p *pumlD2) declare(line string) bool {
	word := pumlFirstWord(strings.ToLower(line))
	shape, ok := pumlDeclShapes[word]
	if !ok {
		return false
	}
	rest := strings.TrimSpace(line[len(word):])
	if pumlStartsWithArrow(rest) {
		return false
	}
	if _, _, _, _, isRel := splitPumlArrow(rest); isRel {
		// `component A --> B` is an edge, not a declaration of "A --> B".
		return false
	}
	name, alias := splitPumlAlias(rest)
	name = strings.TrimSpace(strings.TrimSuffix(name, "{"))
	if !pumlHasAlnum(name) {
		return false
	}
	id := alias
	if id == "" {
		id = name
	}
	p.emitNode(p.idFor(id), name, shape, true)
	return true
}

func (p *pumlD2) handleNote(lines []string, i int) (int, bool) {
	line := strings.TrimSpace(lines[i])
	word := pumlFirstWord(strings.ToLower(line))
	switch word {
	case "note", "hnote", "rnote":
	default:
		return i, false
	}
	rest := strings.TrimSpace(line[len(word):])
	text := ""
	if idx := strings.Index(rest, ":"); idx >= 0 {
		text = strings.TrimSpace(rest[idx+1:])
	} else {
		var parts []string
		j := i + 1
		for ; j < len(lines); j++ {
			t := strings.TrimSpace(lines[j])
			if strings.EqualFold(t, "end note") || strings.EqualFold(t, "endnote") {
				j++
				break
			}
			parts = append(parts, t)
		}
		text = strings.Join(parts, "\n")
		i = j - 1
	}
	p.gen++
	p.emitNode(fmt.Sprintf("note%d", p.gen), text, "callout", true)
	return i, true
}

func (p *pumlD2) handleClass(lines []string, i int) (int, bool) {
	name, alias, block, ok := pumlClassDecl(strings.TrimSpace(lines[i]))
	if !ok {
		return i, false
	}
	var members []string
	if block {
		j := i + 1
		for ; j < len(lines); j++ {
			t := strings.TrimSpace(lines[j])
			if t == "}" {
				j++
				break
			}
			members = append(members, t)
		}
		i = j - 1
	}
	id := alias
	if id == "" {
		id = name
	}
	id = p.idFor(id)
	if len(members) == 0 {
		p.emitNode(id, name, "class", true)
		return i, true
	}
	label := name + "\n" + strings.Join(members, "\n")
	p.emit(id + ": {")
	p.indent++
	p.emit("label: " + pumlLabel(label))
	p.indent--
	p.emit("}")
	return i, true
}

func pumlClassDecl(line string) (name, alias string, block, ok bool) {
	prefixes := []string{
		"abstract class ",
		"abstract ",
		"class ",
		"interface ",
		"enum ",
		"struct ",
		"annotation ",
	}
	low := strings.ToLower(line)
	rest := ""
	for _, prefix := range prefixes {
		if strings.HasPrefix(low, prefix) {
			rest = strings.TrimSpace(line[len(prefix):])
			break
		}
	}
	if rest == "" {
		return "", "", false, false
	}
	if idx := strings.Index(rest, "<"); idx > 0 {
		rest = strings.TrimSpace(rest[:idx])
	}
	if strings.HasSuffix(rest, "{") {
		block = true
		rest = strings.TrimSpace(strings.TrimSuffix(rest, "{"))
	}
	name, alias = splitPumlAlias(rest)
	if name == "" {
		return "", "", false, false
	}
	return name, alias, block, true
}

func (p *pumlD2) handleRelation(line string) bool {
	left, arrow, right, label, ok := splitPumlArrow(line)
	if !ok {
		return false
	}
	left = stripPumlDecoration(cleanPumlName(left))
	right = stripPumlDecoration(cleanPumlName(right))
	if left == "" || right == "" {
		return false
	}
	defLabel, reverse := pumlArrowLabel(arrow)
	if label == "" {
		label = defLabel
	}
	src, dst := left, right
	if reverse {
		src, dst = right, left
	}
	srcID, dstID := p.endpoint(src), p.endpoint(dst)
	if label != "" {
		p.emit(srcID + " -> " + dstID + ": " + pumlLabel(label))
	} else {
		p.emit(srcID + " -> " + dstID)
	}
	return true
}

func (p *pumlD2) openBlock(kind, label string) {
	label = strings.TrimSpace(label)
	name := ""
	if label != "" {
		name = pumlSlug(label)
	}
	if name == "" {
		p.gen++
		name = fmt.Sprintf("%s%d", kind, p.gen)
	}
	id := name
	for n := 2; p.blockIDs[id]; n++ {
		id = fmt.Sprintf("%s_%d", name, n)
	}
	p.blockIDs[id] = true
	p.emit(id + ": {")
	p.indent++
	if label != "" && label != id {
		p.emit("label: " + pumlLabel(label))
	}
	p.blocks = append(p.blocks, id)
}

func (p *pumlD2) closeBlock() {
	if len(p.blocks) == 0 {
		return
	}
	p.blocks = p.blocks[:len(p.blocks)-1]
	if p.indent > 0 {
		p.indent--
	}
	p.emit("}")
}

func (p *pumlD2) closeAll() {
	for len(p.blocks) > 0 {
		p.closeBlock()
	}
}

func (p *pumlD2) idFor(name string) string {
	key := cleanPumlName(name)
	if key == "" {
		key = "node"
	}
	if id, ok := p.ids[key]; ok {
		return id
	}
	base := key
	if !pumlSafeID(base) {
		base = pumlSlug(key)
	}
	id := base
	for n := 2; ; n++ {
		if other, exists := p.taken[id]; !exists || other == key {
			break
		}
		id = fmt.Sprintf("%s_%d", base, n)
	}
	p.ids[key] = id
	p.taken[id] = key
	return id
}

func (p *pumlD2) emitNode(id, label, shape string, standalone bool) {
	n, ok := p.nodes[id]
	if !ok {
		n = &pumlNode{}
		p.nodes[id] = n
	}
	label = strings.TrimSpace(label)
	if label != "" && label != id && label != n.label {
		p.emit(id + ": " + pumlLabel(label))
		n.label = label
	} else if standalone && !ok {
		p.emit(id)
	}
	if shape != "" && n.shape != shape {
		p.emit(id + ".shape: " + shape)
		n.shape = shape
	}
}

func (p *pumlD2) endpoint(name string) string {
	raw := cleanPumlName(name)
	if raw == "" {
		raw = "node"
	}
	id := p.idFor(raw)
	if raw != id {
		p.emitNode(id, raw, "", false)
	}
	return id
}

// ---------------------------------------------------------------------------
// PlantUML -> Mermaid
// ---------------------------------------------------------------------------

type pumlGanttTask struct {
	name      string
	ref       string // Mermaid id, e.g. t1 or task1
	start     string
	afterName string
	duration  string
	section   string
}

func pumlGanttToMermaid(lines []string) (string, error) {
	projectStart := ""
	title := ""
	section := ""
	var tasks []*pumlGanttTask
	byName := make(map[string]*pumlGanttTask)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			continue
		case pumlGanttProjectRe.MatchString(line):
			projectStart = pumlGanttProjectRe.FindStringSubmatch(line)[1]
			continue
		case pumlGanttSectionRe.MatchString(line):
			section = strings.TrimSpace(pumlGanttSectionRe.FindStringSubmatch(line)[1])
			continue
		}
		low := strings.ToLower(line)
		if low == "title" || strings.HasPrefix(low, "title ") || strings.HasPrefix(low, "title:") {
			title = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line[len("title"):]), ":"))
			continue
		}
		if !pumlGanttTaskRe.MatchString(line) {
			continue
		}
		m := pumlGanttTaskRe.FindStringSubmatch(line)
		t := &pumlGanttTask{name: strings.TrimSpace(m[1]), section: section}
		verb := strings.TrimSpace(m[2])
		if d := pumlGanttLastsRe.FindStringSubmatch(verb); d != nil {
			if d[1] != "" {
				t.duration = pumlGanttWeeks(d[1])
			} else {
				t.duration = d[2] + "d"
			}
		}
		if st := pumlGanttStartsRe.FindStringSubmatch(verb); st != nil {
			t.start = st[1]
		}
		if a := pumlGanttAfterRe.FindStringSubmatch(verb); a != nil {
			t.afterName = strings.TrimSpace(a[1])
		}
		tasks = append(tasks, t)
		if _, ok := byName[t.name]; !ok {
			byName[t.name] = t
		}
	}

	out := []string{"gantt", "  dateFormat YYYY-MM-DD"}
	if title != "" {
		out = append(out, "  title "+title)
	}

	// Mermaid task ids are either explicit (`t1`) or auto-generated (`task1`,
	// counted over tasks without an explicit id); mirror both so `after`
	// references always resolve.
	numbered, auto := 0, 0
	prev := ""
	current := ""
	for _, t := range tasks {
		if t.section != current {
			if t.section != "" {
				out = append(out, "  section "+t.section)
			}
			current = t.section
		}
		start := t.start
		if start == "" {
			if dep, ok := byName[t.afterName]; ok {
				start = "after " + dep.ref
			} else if projectStart != "" {
				start = projectStart
			} else if prev != "" {
				start = "after " + prev
			}
		}
		duration := t.duration
		if duration == "" {
			duration = "1d"
		}
		if start == "" {
			auto++
			t.ref = fmt.Sprintf("task%d", auto)
			out = append(out, "  "+t.name+" :"+duration)
		} else {
			numbered++
			t.ref = fmt.Sprintf("t%d", numbered)
			out = append(out, "  "+t.name+" :"+t.ref+", "+start+", "+duration)
		}
		prev = t.ref
	}
	return strings.Join(out, "\n") + "\n", nil
}

func pumlGanttWeeks(weeks string) string {
	n := 0
	for _, r := range weeks {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	if n == 0 {
		n = 1
	}
	return fmt.Sprintf("%dd", n*7)
}

func pumlMindmapToMermaid(lines []string) (string, error) {
	out := []string{"mindmap"}
	rootSeen := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		depth, text := pumlMindmapNode(line)
		if depth == 0 || text == "" {
			continue
		}
		if depth == 1 && !rootSeen {
			rootSeen = true
			out = append(out, "  root(("+pumlMindmapRootText(text)+"))")
			continue
		}
		if depth == 1 {
			// PlantUML side branches become children of the single Mermaid root.
			depth = 2
		}
		out = append(out, strings.Repeat("  ", depth)+text)
	}
	return strings.Join(out, "\n") + "\n", nil
}

// pumlMindmapNode returns the depth of a mindmap line (the run of *, + and -
// side markers) and its text.
func pumlMindmapNode(line string) (int, string) {
	depth := 0
	for depth < len(line) {
		switch line[depth] {
		case '*', '+', '-':
			depth++
		default:
			return depth, strings.TrimSpace(line[depth:])
		}
	}
	return depth, ""
}

func pumlMindmapRootText(text string) string {
	repl := strings.NewReplacer("(", "", ")", "", "[", "", "]", "", "{", "", "}", "", `"`, "")
	if s := strings.TrimSpace(repl.Replace(text)); s != "" {
		return s
	}
	return "root"
}

// ---------------------------------------------------------------------------
// Line splitting helpers
// ---------------------------------------------------------------------------

func pumlIsArrowByte(c byte) bool {
	switch c {
	case '-', '.', '<', '>', '|', '*', 'o', 'x', '+':
		return true
	}
	return false
}

// splitPumlArrow splits `left <arrow> right : label` into its parts.
func splitPumlArrow(line string) (left, arrow, right, label string, ok bool) {
	start, end := -1, -1
	for i := 0; i < len(line); i++ {
		if !pumlIsArrowByte(line[i]) {
			continue
		}
		j := i
		for j < len(line) && pumlIsArrowByte(line[j]) {
			j++
		}
		if run := line[i:j]; len(run) >= 2 && strings.ContainsAny(run, "-.") {
			start, end = i, j
			break
		}
		i = j
	}
	if start <= 0 || end >= len(line) {
		return "", "", "", "", false
	}
	left = strings.TrimSpace(line[:start])
	arrow = line[start:end]
	rest := strings.TrimSpace(line[end:])
	if k := strings.Index(rest, ":"); k >= 0 {
		label = strings.TrimSpace(rest[k+1:])
		rest = strings.TrimSpace(rest[:k])
	}
	rest = stripPumlDecoration(rest)
	if left == "" || rest == "" {
		return "", "", "", "", false
	}
	return left, arrow, rest, label, true
}

func pumlArrowLabel(arrow string) (label string, reverse bool) {
	switch {
	case strings.HasPrefix(arrow, "<|"):
		return "inherits", true
	case strings.HasPrefix(arrow, "<"):
		return "", true
	case strings.HasPrefix(arrow, "*"):
		return "composition", false
	case strings.HasPrefix(arrow, "o"):
		return "aggregation", false
	case strings.HasPrefix(arrow, "..>"):
		return "depends on", false
	}
	return "", false
}

func stripPumlDecoration(s string) string {
	s = strings.TrimSpace(s)
	for {
		changed := false
		for _, suffix := range []string{"++", "--", "**", "!!", "??", "<<", ">>"} {
			if strings.HasSuffix(s, suffix) {
				s = strings.TrimSpace(strings.TrimSuffix(s, suffix))
				changed = true
			}
		}
		if !changed {
			return s
		}
	}
}

func splitPumlAlias(s string) (name, alias string) {
	s = strings.TrimSpace(s)
	if m := pumlAsRe.FindStringSubmatch(s); m != nil {
		return cleanPumlName(m[1]), strings.TrimSpace(m[2])
	}
	return cleanPumlName(s), ""
}

func cleanPumlName(s string) string {
	s = strings.TrimSpace(s)
	for {
		changed := false
		if len(s) >= 2 && strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			s, changed = strings.TrimSpace(s[1:len(s)-1]), true
		}
		if len(s) >= 2 && strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
			s, changed = strings.TrimSpace(s[1:len(s)-1]), true
		}
		if len(s) >= 2 && strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			s, changed = strings.TrimSpace(s[1:len(s)-1]), true
		}
		if !changed {
			return s
		}
	}
}

// pumlStartsWithArrow reports whether s opens with an arrow such as `-->`.
func pumlStartsWithArrow(s string) bool {
	i := 0
	for i < len(s) && pumlIsArrowByte(s[i]) {
		i++
	}
	if i < 2 {
		return false
	}
	return strings.ContainsAny(s[:i], "-.")
}

// pumlHasAlnum reports whether s holds at least one letter or digit, which
// keeps punctuation-only fragments from becoming nodes.
func pumlHasAlnum(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return true
		}
	}
	return false
}

func pumlSafeID(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r == '_' ||
			(r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

func pumlSlug(s string) string {
	var b strings.Builder
	underscore := false
	for _, r := range s {
		if r == '_' ||
			(r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') {
			b.WriteRune(r)
			underscore = false
			continue
		}
		if b.Len() > 0 && !underscore {
			b.WriteByte('_')
			underscore = true
		}
	}
	if out := strings.Trim(b.String(), "_"); out != "" {
		return out
	}
	return "node"
}

// pumlLabel renders a label, quoting only when the raw text is ambiguous in D2.
func pumlLabel(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, "\"{}[];|#\\\n\r\t") {
		return pumlQuote(s)
	}
	return s
}

func pumlQuote(s string) string {
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
