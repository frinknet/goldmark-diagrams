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
	return &RenderResult{SVG: svg, Engine: EnginePele, Width: w, Height: h}, nil
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
