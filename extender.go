package diagrams

import (
	"bytes"
	"context"
	"strings"

	"github.com/goldmark-diagrams/goldmark-diagrams/engines"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// transformPriority places the diagram transformer after the core block
// parsers have produced their fenced code blocks.
const transformPriority = 100

// diagramKind is the custom AST node kind that carries pre-rendered HTML.
var diagramKind = ast.NewNodeKind("GoldmarkDiagram")

// diagramBlock holds one rendered <figure> fragment in the document tree.
type diagramBlock struct {
	ast.BaseBlock
	lang string
	html string
	err  error
}

// Kind implements ast.Node.
func (n *diagramBlock) Kind() ast.NodeKind { return diagramKind }

// IsRaw implements ast.Node: the fragment is emitted verbatim.
func (n *diagramBlock) IsRaw() bool { return true }

// Dump implements ast.Node.
func (n *diagramBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Lang": n.lang}, nil)
}

// extender is the goldmark.Extender returned by New.
type extender struct {
	config *Config
}

// Extend registers the AST transformer and the HTML node renderer.
func (e *extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&transformer{config: e.config}, transformPriority),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&diagramRenderer{}, transformPriority),
	))
}

// transformer replaces supported fenced code blocks with a diagramBlock.
type transformer struct {
	config *Config
}

// Transform implements parser.ASTTransformer.
func (t *transformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	var fences []*ast.FencedCodeBlock
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if fcb, ok := n.(*ast.FencedCodeBlock); ok {
			fences = append(fences, fcb)
		}
		return ast.WalkContinue, nil
	})
	if len(fences) == 0 {
		return
	}

	for _, fcb := range fences {
		if !t.config.handles(string(fcb.Language(source))) {
			continue
		}
		lang := engines.NormalizeLang(string(fcb.Language(source)))
		attrs := parseInfoAttrs(string(fcb.Info.Text(source)))
		body := fencedSource(fcb, source)

		block := &diagramBlock{lang: lang}
		block.html, block.err = renderFragment(context.Background(), t.config, lang, body, attrs, fenceLine(source, fcb))

		parent := fcb.Parent()
		if parent == nil {
			continue
		}
		parent.ReplaceChild(parent, fcb, block)
	}
}

// handles reports whether a fence language is intercepted by this config.
func (c *Config) handles(lang string) bool {
	canonical := engines.NormalizeLang(lang)
	if canonical == "" {
		return false
	}
	for _, l := range c.Languages {
		if engines.NormalizeLang(l) == canonical {
			return true
		}
	}
	return false
}

// diagramRenderer writes the pre-rendered fragment for each diagramBlock.
type diagramRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *diagramRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(diagramKind, r.render)
}

func (r *diagramRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	block, ok := node.(*diagramBlock)
	if !ok {
		return ast.WalkSkipChildren, nil
	}
	if block.err != nil {
		// ErrorFail: surface the engine error from markdown.Convert().
		return ast.WalkStop, block.err
	}
	if _, err := w.WriteString(block.html); err != nil {
		return ast.WalkStop, err
	}
	return ast.WalkSkipChildren, nil
}

// fencedSource joins the fenced block's lines into the engine source text.
func fencedSource(fcb *ast.FencedCodeBlock, source []byte) string {
	lines := fcb.Lines()
	if lines == nil {
		return ""
	}
	var b bytes.Buffer
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		b.Write(seg.Value(source))
	}
	return b.String()
}

// fenceLine returns the 1-based source line of the opening fence, or 0 when the
// position is not recoverable.
func fenceLine(source []byte, fcb *ast.FencedCodeBlock) int {
	lines := fcb.Lines()
	if lines == nil || lines.Len() == 0 {
		return 0
	}
	start := lines.At(0).Start
	if start <= 0 || start > len(source) {
		return 0
	}
	// The first content line is one line below the opening fence.
	return bytes.Count(source[:start], []byte{'\n'})
}

// parseInfoAttrs reads the key=value attributes that follow the fence language,
// e.g. "mermaid id=brd_123 theme=dark". Values may be double quoted.
func parseInfoAttrs(info string) map[string]string {
	fields := splitInfo(info)
	if len(fields) <= 1 {
		return nil
	}
	attrs := make(map[string]string, len(fields)-1)
	for _, f := range fields[1:] {
		k, v, ok := strings.Cut(f, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		attrs[k] = strings.Trim(v, `"`)
	}
	if len(attrs) == 0 {
		return nil
	}
	return attrs
}

// splitInfo splits an info string on whitespace, keeping double-quoted values
// (which may contain spaces) in one field.
func splitInfo(info string) []string {
	var fields []string
	var cur strings.Builder
	inQuote := false
	for _, r := range info {
		switch {
		case r == '"':
			inQuote = !inQuote
			cur.WriteRune(r)
		case !inQuote && (r == ' ' || r == '\t' || r == '\n' || r == '\r'):
			if cur.Len() > 0 {
				fields = append(fields, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		fields = append(fields, cur.String())
	}
	return fields
}
