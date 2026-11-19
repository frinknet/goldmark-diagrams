package engines

import (
	"context"
	"strings"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

// renderD2 compiles D2 source with the pure-Go dagre layout engine.
// "elk" is intentionally unsupported: it requires an external plugin binary,
// which would violate the CGO_ENABLED=0 / zero-CLI guarantee.
func renderD2(ctx context.Context, meta Meta, source string) (*RenderResult, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, err
	}
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2dagrelayout.DefaultLayout, nil
	}

	renderOpts := &d2svg.RenderOpts{
		Pad:      go2.Pointer(int64(10)),
		Salt:     go2.Pointer("goldmark-diagrams:" + meta.Lang),
		NoXMLTag: go2.Pointer(true),
	}
	switch meta.Theme {
	case "dark":
		renderOpts.ThemeID = go2.Pointer(int64(d2themescatalog.DarkMauve.ID))
		renderOpts.DarkThemeID = go2.Pointer(int64(d2themescatalog.DarkMauve.ID))
	case "neutral":
		renderOpts.ThemeID = go2.Pointer(int64(d2themescatalog.NeutralGrey.ID))
		renderOpts.DarkThemeID = go2.Pointer(int64(d2themescatalog.DarkMauve.ID))
	default: // auto / light
		renderOpts.ThemeID = go2.Pointer(int64(d2themescatalog.NeutralDefault.ID))
		renderOpts.DarkThemeID = go2.Pointer(int64(d2themescatalog.DarkMauve.ID))
	}

	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Layout:         go2.Pointer("dagre"),
		Ruler:          ruler,
		InputPath:      meta.Lang + ".d2",
	}

	// The D2 layout engine writes its own diagnostics to the context logger.
	dctx := log.WithDefault(ctx)
	if dctx == nil {
		dctx = context.Background()
	}

	// D2's compiler has no internal cancellation checkpoints, so a pathological
	// source can outlive the caller's deadline. Compile in a goroutine and
	// return as soon as the context ends; the abandoned goroutine still runs to
	// completion but no longer blocks the caller.
	type d2Result struct {
		svg string
		err error
	}
	done := make(chan d2Result, 1)
	go func() {
		diagram, _, err := d2lib.Compile(dctx, source, compileOpts, renderOpts)
		if err != nil {
			done <- d2Result{err: err}
			return
		}
		out, err := d2svg.Render(diagram, renderOpts)
		if err != nil {
			done <- d2Result{err: err}
			return
		}
		done <- d2Result{svg: string(out)}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-done:
		if r.err != nil {
			return nil, r.err
		}
		svg := r.svg
		if i := strings.Index(svg, "<svg"); i >= 0 {
			svg = svg[i:]
		} else {
			return nil, &engineError{"d2 produced no <svg> element"}
		}
		var w, h float64
		if vb := parseViewBox(svg); vb != nil {
			w, h = vb[2], vb[3]
		}
		return &RenderResult{SVG: svg, Engine: EngineD2, Width: w, Height: h}, nil
	}
}
