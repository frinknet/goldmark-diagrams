// Package diagrams is a 100% pure-Go Goldmark extension that compiles fenced
// diagram and chart code blocks into one standardized, responsive,
// CSS-interactive <svg> output.
//
// Engines (all pure Go, CGO_ENABLED=0 safe):
//   - pele-go (Mermaid, all 33 diagram types) for ```mermaid / ```mmd
//   - D2 (dagre layout) for ```d2, and as target for the PlantUML / DOT transpilers
//   - go-analyze/charts (SVG painter) for ```chart / ```echarts
//
// Hard guarantees:
//   - Never emits <canvas>, <img>, or inline <script>.
//   - Never fails markdown conversion: errors render as an inline styled SVG card.
//   - Every SVG passes through NormalizeSVG before caching and HTML emission.
package diagrams

import (
	"context"
	_ "embed"
	"time"

	"github.com/goldmark-diagrams/goldmark-diagrams/pkg/cache"
	"github.com/goldmark-diagrams/goldmark-diagrams/pkg/engines"
	"github.com/goldmark-diagrams/goldmark-diagrams/pkg/extender"
	"github.com/goldmark-diagrams/goldmark-diagrams/pkg/render"
	"github.com/yuin/goldmark"
)

// ModuleVersion participates in the content-addressed cache key, so bumping it
// invalidates every cached fragment after a renderer change.
const ModuleVersion = cache.ModuleVersion

//go:embed assets/diagrams.css
var defaultCSS string

// DefaultCSS returns the reference pure-CSS interactivity and theme stylesheet.
// Serve it from your own route (for example /_diagrams/css).
func DefaultCSS() string { return defaultCSS }

// SupportedLanguages lists every fence language the extension recognizes.
func SupportedLanguages() []string { return engines.SupportedLanguages() }

// ErrorMode selects what happens when a diagram fails to compile.
type ErrorMode = render.ErrorMode

const (
	// ErrorSVGCard (default) renders a valid inline <svg> error card.
	ErrorSVGCard = render.ErrorSVGCard
	// ErrorCodeBlock falls back to a standard <pre><code class="language-..."> block.
	ErrorCodeBlock = render.ErrorCodeBlock
	// ErrorFail returns the Go error from markdown.Convert().
	ErrorFail = render.ErrorFail
)

// Cache is a content-addressed store of normalized <figure> HTML fragments.
type Cache = cache.Cache

// Config holds every tunable of the extension. It is exported so hosts can
// inspect a resolved configuration, but should be built via Options.
type Config = render.Config

// Option configures New() and RenderStandalone().
type Option = render.Option

// WithTheme sets the theme hint ("auto" | "light" | "dark" | "neutral") emitted
// as data-theme on <figure> and mapped onto engine theme selection.
func WithTheme(theme string) Option { return func(c *Config) { c.Theme = theme } }

// WithCache replaces the default in-memory LRU cache.
func WithCache(cache Cache) Option { return func(c *Config) { c.Cache = cache } }

// WithErrorMode selects the failure behavior. Default: ErrorSVGCard.
func WithErrorMode(mode ErrorMode) Option { return func(c *Config) { c.ErrorMode = mode } }

// WithLanguages restricts which fence languages are intercepted.
// Default: all supported languages (see engines.SupportedLanguages).
func WithLanguages(langs ...string) Option { return func(c *Config) { c.Languages = langs } }

// WithTimeout bounds a single engine render. Default: 5 * time.Second.
func WithTimeout(d time.Duration) Option { return func(c *Config) { c.Timeout = d } }

// WithD2Layout selects the D2 layout engine. Only "dagre" is pure Go; any
// other value (e.g. "elk", which shells out to a plugin binary) is ignored
// with dagre kept, preserving the zero-external-binary guarantee.
func WithD2Layout(layout string) Option { return func(c *Config) { c.D2Layout = layout } }

// WithID stamps data-id on the <figure> element for standalone renders,
// enabling server-backed feedback loops (kanban moves, chart drill-downs).
func WithID(id string) Option { return func(c *Config) { c.ID = id } }

// NewMemoryLRU returns a thread-safe in-memory LRU cache holding at most
// maxEntries fragments. A non-positive maxEntries falls back to 512.
func NewMemoryLRU(maxEntries int) Cache { return cache.NewMemoryLRU(maxEntries) }

// NewDiskCache returns a cache that stores fragments as <hash>.svg files under
// dir, writing each atomically.
func NewDiskCache(dir string) Cache { return cache.NewDiskCache(dir) }

// New returns a goldmark.Extender for use with goldmark.New(goldmark.WithExtensions(...)).
func New(opts ...Option) goldmark.Extender {
	return extender.New(render.Default(opts...))
}

// RenderStandalone compiles a single diagram directly to normalized
// <figure><svg> HTML. Backends use it to re-render interactive Kanban/Gantt
// SVG fragments after database mutations, so the client can swap
// figure.outerHTML with zero state drift.
func RenderStandalone(ctx context.Context, lang, source string, opts ...Option) (string, error) {
	return render.Fragment(ctx, render.Default(opts...), lang, source, nil, 0)
}

// NormalizeSVG rewrites a raw engine SVG into the single <svg> output contract:
// a responsive viewBox, no hardcoded pixel size, the svg-diagram-svg class, and
// data-* hooks on the interactive elements. It never emits <canvas>, <img>, or
// inline <script>.
func NormalizeSVG(rawSVG string, meta engines.Meta) (string, error) {
	return render.NormalizeSVG(rawSVG, meta)
}
