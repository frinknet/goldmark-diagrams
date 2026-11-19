package engines

import (
	"context"
	"strconv"
	"strings"

	pele "github.com/Broderick-Westrope/pele-go"
	_ "github.com/Broderick-Westrope/pele-go/all" // register all 33 Mermaid diagram types
)

// renderPele compiles Mermaid source with pele-go — a pure-Go port covering
// kanban, mindmap, gantt, venn, gitGraph, flowchart, sequence, ERD, class,
// state, C4, sankey, radar, timeline, quadrant, pie, xychart and more.
func renderPele(ctx context.Context, meta Meta, source string) (*RenderResult, error) {
	opts := pele.Options{}
	// Theme hint becomes Mermaid configuration; pele reads it from Options.Config.
	// Mermaid's own theme names are default, forest, dark, neutral and base.
	switch meta.Theme {
	case "dark":
		if cfg, err := pele.ParseConfig(`{"theme":"dark"}`); err == nil {
			opts.Config = cfg
		}
	case "neutral":
		if cfg, err := pele.ParseConfig(`{"theme":"neutral"}`); err == nil {
			opts.Config = cfg
		}
	case "light":
		if cfg, err := pele.ParseConfig(`{"theme":"default"}`); err == nil {
			opts.Config = cfg
		}
	}

	res, err := pele.Render(ctx, source, opts)
	if err != nil {
		return nil, err
	}
	svg := res.SVG
	// Strip the XML prolog if present; we inline SVG into HTML.
	if i := strings.Index(svg, "<svg"); i > 0 {
		svg = svg[i:]
	} else if i < 0 {
		return nil, errNoSVG
	}
	w, h := res.Width, res.Height
	if w == 0 || h == 0 {
		// Fall back to viewBox parsing when the engine didn't report dimensions.
		if vb := parseViewBox(svg); vb != nil {
			w, h = vb[2], vb[3]
		}
	}
	return &RenderResult{SVG: svg, Engine: EnginePele, Labels: xychartLabels(source), Width: w, Height: h}, nil
}

// xychartLabels extracts the category names of an xychart-beta x-axis band so
// the normalizer can name the bars and points pele draws without an id. A
// numeric axis range has no categories and yields nil, leaving marks numbered.
func xychartLabels(source string) []string {
	isXY := false
	for _, line := range strings.Split(source, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "xychart-beta") {
			isXY = true
			break
		}
	}
	if !isXY {
		return nil
	}
	for _, line := range strings.Split(source, "\n") {
		axis := strings.TrimSpace(line)
		if !strings.HasPrefix(axis, "x-axis") {
			continue
		}
		open := strings.IndexByte(axis, '[')
		closeAt := strings.LastIndexByte(axis, ']')
		if open < 0 || closeAt <= open {
			continue
		}
		return splitAxisLabels(axis[open+1 : closeAt])
	}
	return nil
}

// splitAxisLabels splits a bracket body on commas that are outside quotes, so a
// quoted category may contain one.
func splitAxisLabels(body string) []string {
	var labels []string
	var cur strings.Builder
	var quote rune
	for _, r := range body {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
				continue
			}
			cur.WriteRune(r)
		case r == '"' || r == '\'':
			quote = r
		case r == ',':
			labels = append(labels, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	return append(labels, strings.TrimSpace(cur.String()))
}

var errNoSVG = &engineError{"pele produced no <svg> element"}

type engineError struct{ msg string }

func (e *engineError) Error() string { return e.msg }

// parseViewBox extracts "0 0 W H" numbers from an SVG string.
func parseViewBox(svg string) []float64 {
	i := strings.Index(svg, `viewBox="`)
	if i < 0 {
		return nil
	}
	rest := svg[i+len(`viewBox="`):]
	j := strings.Index(rest, `"`)
	if j < 0 {
		return nil
	}
	fields := strings.Fields(rest[:j])
	if len(fields) != 4 {
		return nil
	}
	out := make([]float64, 4)
	for k, f := range fields {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return nil
		}
		out[k] = v
	}
	return out
}
