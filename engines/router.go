// Package engines dispatches fenced diagram sources to one of three pure-Go
// rendering engines: pele-go (Mermaid), D2 (dagre), or go-analyze/charts.
package engines

import (
	"context"
	"fmt"
	"strings"

	"github.com/goldmark-diagrams/goldmark-diagrams/engines/transpilers"
)

// Engine names reported on <figure data-engine>.
const (
	EnginePele   = "pele"
	EngineD2     = "d2"
	EngineCharts = "charts"
)

// Meta describes a render request and the engine that handled it.
type Meta struct {
	Lang     string            // normalized fence language, e.g. "mermaid", "d2", "chart"
	Engine   string            // EnginePele | EngineD2 | EngineCharts
	Theme    string            // "auto" | "light" | "dark" | "neutral"
	D2Layout string            // resolved D2 layout ("dagre")
	Attrs    map[string]string // key=value attributes from the info string (id=..., resource=...)
	Line     int               // source line of the fence, when known
	Labels   []string          // category labels, when the source declares them (charts)

	// Width and Height are the intrinsic drawing size, filled in after a
	// successful render so the normalizer can synthesize a viewBox.
	Width  float64
	Height float64
}

// RenderResult is the raw SVG plus intrinsic size produced by an engine.
type RenderResult struct {
	SVG    string
	Engine string   // EnginePele | EngineD2 | EngineCharts
	Labels []string // category labels the normalizer can name marks with
	Width  float64  // 0 = unknown; normalizer will try to read viewBox
	Height float64
}

// SupportedLanguages lists every fence language the router recognizes.
func SupportedLanguages() []string {
	return []string{"mermaid", "mmd", "d2", "plantuml", "puml", "dot", "graphviz", "chart", "echarts"}
}

// NormalizeLang lowercases and maps aliases onto canonical languages.
// It returns "" for anything unsupported.
func NormalizeLang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "mermaid", "mmd":
		return "mermaid"
	case "d2":
		return "d2"
	case "plantuml", "puml":
		return "plantuml"
	case "dot", "graphviz":
		return "dot"
	case "chart", "echarts":
		return "chart"
	}
	return ""
}

// EngineForLang reports the engine a canonical language routes to, so callers
// can label output before a render succeeds. PlantUML is reported as D2 because
// only its gantt and mindmap forms are drawn by pele-go at render time.
func EngineForLang(lang string) string {
	switch NormalizeLang(lang) {
	case "mermaid":
		return EnginePele
	case "d2", "plantuml", "dot":
		return EngineD2
	case "chart":
		return EngineCharts
	}
	return ""
}

// Dispatch routes source through preprocessing into the correct engine.
func Dispatch(ctx context.Context, meta Meta, source string) (*RenderResult, error) {
	switch meta.Lang {
	case "mermaid":
		return renderPele(ctx, meta, source)
	case "d2":
		return renderD2(ctx, meta, source)
	case "plantuml":
		// D2 cannot draw gantt or mindmap; those forms transpile to Mermaid and
		// go to pele-go, which covers all 33 Mermaid types. Everything else
		// goes through D2.
		if mermaid, err := transpilers.PlantUMLToMermaid(source); err == nil {
			return renderPele(ctx, meta, mermaid)
		}
		d2src, err := transpilers.PlantUMLToD2(source)
		if err != nil {
			return nil, fmt.Errorf("plantuml transpile: %w", err)
		}
		return renderD2(ctx, meta, d2src)
	case "dot":
		d2src, err := transpilers.DotToD2(source)
		if err != nil {
			return nil, fmt.Errorf("dot transpile: %w", err)
		}
		return renderD2(ctx, meta, d2src)
	case "chart":
		return renderCharts(ctx, meta, source)
	}
	return nil, fmt.Errorf("unsupported diagram language %q", meta.Lang)
}
