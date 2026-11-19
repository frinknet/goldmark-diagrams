Engineering Specification: goldmark-diagrams

Here is the complete specification for goldmark-diagrams incorporating the 3-engine pure-Go stack (pele-go/all + d2 + go-analyze/charts), zero WASM/CGo guarantees, unified SVG normalization, pure-CSS interactivity, and data-* hooks for server-side feedback loops.


1. Purpose & Core Constraints

goldmark-diagrams is a 100% pure-Go Goldmark extension (goldmark.Extender) designed for embedding inside custom Caddy builds, documentation engines, and Go web applications. It compiles fenced diagram and chart code blocks into one standardized, responsive, CSS-interactive <svg> output with optional data-* hooks for server-backed interactivity (such as chart drill-downs or other diagram mutations).

Hard Engineering Rules

1. Pure Go Only (CGO_ENABLED=0): Zero CGo, zero WASM/wazero runtimes, zero external CLI binaries (mmdc, plantuml.jar, dot), and zero headless browsers or Node/Java runtimes.
2. Single <svg> Output Contract: Every supported fenced block renders into a normalized <figure class="svg-diagram"> wrapping an inline <svg viewBox="...">. Never emit <canvas>, <img>, or inline <script> tags.
3. Never-Fail Markdown Conversion: Invalid diagram syntax or layout timeouts must never break markdown.Convert(). Errors render into an inline styled <svg> error card by default.
4. CSS-First Interactivity + Server-SVG Feedback: SVGs emit predictable data-* attributes and CSS custom properties (--diagram-*) so hover dimming, animated flow lines, and dark/light themes work in pure CSS, while thin client-side listeners can round-trip mutations back to the server and swap in a freshly rendered SVG fragment.

2. Go Environment & Dependencies

• Go Version: Go 1.26+ (required by pele-go)
• Module path: github.com/goldmark-diagrams/goldmark-diagrams
• Allowed Dependencies:
	• github.com/yuin/goldmark — Markdown AST & HTML rendering
	• github.com/Broderick-Westrope/pele-go (importing _ "github.com/Broderick-Westrope/pele-go/all") — Pure-Go Mermaid engine supporting all 33 Mermaid diagram types (Kanban, Mindmap, Gantt, Venn, GitGraph, Flowchart, Sequence, ERD, Class, State, C4, Sankey, Radar, Timeline, Quadrant, etc.)
	• oss.terrastruct.com/d2 — Pure-Go D2 compiler & layout engine (dagre only; elk/taro require an external binary and are rejected)
	• github.com/go-analyze/charts — Pure-Go SVG charting library, rendered as SVG through charts.SVGOutputOptionFunc() (or PainterOptions.OutputFormat = charts.ChartOutputSVG); there is no charts.SVGOutput() function
	• gopkg.in/yaml.v3 — Optional YAML parsing for chart specs
	• Go Standard Library (encoding/xml, crypto/sha256, strings, regexp, sync, time)

3. Fenced Code Routing Matrix

goldmark-diagrams registers a Goldmark ASTTransformer (priority 100) that intercepts ast.FencedCodeBlock nodes, parses the info string (lang + optional key-value attributes like ```mermaid id=brd_123 theme=dark), and routes the source to one of three pure-Go engines:

Fence Language (```lang)	Parser / Preprocessor	Pure-Go Rendering Engine	Covered Diagram & Chart Types
mermaid, mmd	Direct pass-through	pele-go/all	All 33 Mermaid types: kanban, mindmap, gantt, venn-beta, gitGraph, flowchart, sequenceDiagram, erDiagram, classDiagram, stateDiagram, pie, xychart-beta, quadrantChart, sankey-beta, radar, timeline, c4, journey, etc.
d2	Direct pass-through	oss.terrastruct.com/d2	Architecture, Flow, Sequence, ERD (sql_table), Class, State, Grid, Cloud Icons.
plantuml, puml	transpilers.PlantUMLToD2() / PlantUMLToMermaid()	oss.terrastruct.com/d2, or pele-go for @startgantt/@startmindmap	@startuml Sequence, Class, Component, State, UseCase, Gantt, Mindmap.
dot, graphviz	transpilers.DotToD2()	oss.terrastruct.com/d2	digraph, graph, subgraph cluster_*, rankdir, shapes, and edge labels.
chart, echarts	JSON / YAML unmarshal	go-analyze/charts (SVG painter)	Line, Bar, Stacked Bar, Horizontal Bar, Pie, Donut, Radar, Funnel, Scatter, Candlestick, Heatmap, Table.

Attributes are URI-shaped key=value pairs; each key becomes a sanitized data-* attribute on the <figure> (id=brd_123 emits data-id="brd_123", resource=orders emits data-resource="orders"). Reserved names data-lang, data-engine, data-theme, and data-hash are written by the normalizer and cannot be overridden from a fence.

4. Standardized HTML & SVG Output Contract (normalizer.go)

Every raw <svg> produced by pele-go, d2, or go-analyze/charts must pass through NormalizeSVG(rawSVG string, meta engines.Meta) (string, error) before caching and HTML emission.

Output Structure

html
<figure class="svg-diagram"
        data-lang="mermaid"
        data-engine="pele"
        data-theme="auto"
        data-hash="9f86d081"
        data-id="brd_123">
  <svg viewBox="0 0 900 480" class="svg-diagram-svg" role="img" aria-label="mermaid diagram">
    <!-- Normalized SVG nodes stamped with data-item, data-role, data-series,
         data-column-id, data-card-id, and CSS variables -->
  </svg>
</figure>
Use code with caution.

data-theme normalizes to auto, light, dark, or neutral; any other fence value becomes auto. data-hash is the first eight hex digits of the SHA-256 of the normalized SVG. data-engine is populated before rendering, so an error card still identifies the engine.

Normalization Rules

1. Responsive viewBox: Ensure viewBox="0 0 W H" exists on <svg> (deriving from width/height attributes if absent) and remove hardcoded pixel width/height attributes so CSS governs sizing. The root <svg> carries class svg-diagram-svg, role="img", and an aria-label derived from the fence language when the engine did not set one.
2. Interactive data-* Stamping:
	• Stamp data-item="<id-or-index>" onto primary visual elements (nodes, bars, pie slices, git commits, Gantt tasks). Never invent an index when the element has a name; a chart spec with no labels is the one data set that has none, so its marks carry the 1-based category number the axis prints.
	• Stamp data-role onto the same elements. The role is derived, not enumerated: pele's own class remainder for Mermaid, shape/connection for D2, mark/line for charts, row/cell for table charts. Types that name their own shapes add their own (lifeline, lane, rule, curve, quadrant, legend-item, sankey-node, sankey-link, bar, line, row, cell, version).
	• Echo the role as a class (svg-node, svg-edge, svg-shape, svg-connection, svg-mark, ...), and add svg-series-0..7 to chart and XY-chart marks, so one stylesheet addresses every engine.
	• For Kanban diagrams, stamp data-column-id="<slug>" on column groups and data-card-id="<id>" on card groups so client-side pointer listeners can identify a mutation target.
	• Stamp data-series="<0..N>" onto multi-series groups and paths.
	• Ensure stamped interactive elements include a <title> child for zero-JS native hover tooltips.
3. CSS Custom Property Mapping: Map engine theme fills and strokes to CSS variables with sensible fallbacks:
	• --diagram-bg (default transparent)
	• --diagram-fg (default #1e293b)
	• --diagram-muted (default #64748b)
	• --diagram-border (default #cbd5e1)
	• --diagram-line (default #94a3b8)
	• --diagram-surface (default #ffffff)
	• --diagram-surface-alt (default #f1f5f9)
	• --diagram-accent (default #5470c6)
	• --diagram-series-0 through --diagram-series-7
	• --diagram-error-bg, --diagram-error-border, --diagram-error-fg for error cards

Mermaid's --pele-* variables are aliased onto --diagram-*; D2's embedded palette rules are rewritten by role; chart paint becomes var(--diagram-series-N, <engine colour>). Keep a hook's identity real: Mermaid elements use pele's own ids, a pele connector without an id is named after the two items its path joins (A->B), D2 shapes use their label, D2 connections use their endpoints (alpha->beta, alpha->beta[1]), chart marks use the category label, kanban uses the column/card name, a table row is named after its first cell (header for the header band), and a cell is named after its column header.

5. Server-SVG Feedback Loop

The extension has no database dependency; it only guarantees the hook vocabulary a host needs to build one:
1. Initial Render: The host renders Markdown, or calls diagrams.RenderStandalone(ctx, "mermaid", source, diagrams.WithID("brd_123")), to produce the <figure data-id="brd_123"> SVG.
2. Client Interaction: A thin delegated pointer/click listener reads data-id, data-item, data-column-id, or data-card-id from the clicked/dropped SVG element and POSTs JSON to a host endpoint (gated by the host's own auth middleware).
3. SVG Swap: The host updates its state, re-renders the diagram fragment, and returns it so the client swaps figure.outerHTML = updatedHTML with zero client state drift.

6. Reference Companion CSS (diagrams.DefaultCSS())

Embeddable stylesheet shipped with the package for out-of-the-box pure-CSS interactivity:

css
.svg-diagram {
  margin: 1.5rem 0;
  overflow-x: auto;

  --diagram-bg: transparent;
  --diagram-fg: #1e293b;
  --diagram-muted: #64748b;
  --diagram-border: #cbd5e1;
  --diagram-line: #94a3b8;
  --diagram-surface: #ffffff;
  --diagram-surface-alt: #f1f5f9;
  --diagram-accent: #5470c6;
  --diagram-series-0: #5470c6;
  /* --diagram-series-1 .. --diagram-series-7 */
}

.svg-diagram-svg {
  display: block;
  width: 100%;
  max-width: 100%;
  height: auto;
}

@media (prefers-color-scheme: dark) {
  .svg-diagram {
    --diagram-fg: #e2e8f0;
    --diagram-border: #475569;
    --diagram-surface: #1e293b;
    /* ... */
  }
}

.svg-diagram[data-theme="dark"] {
  --diagram-fg: #e2e8f0;
  --diagram-border: #475569;
  --diagram-surface: #1e293b;
  /* ... */
}

/* Pure-CSS sibling dimming on hover */
.svg-diagram-svg:has([data-item]:hover) [data-item]:not(:hover):not(:has([data-item]:hover)) {
  opacity: 0.35;
  transition: opacity 150ms ease;
}

.svg-diagram-svg [data-item]:hover,
.svg-diagram-svg [data-card-id]:hover {
  filter: brightness(1.08);
  cursor: pointer;
}

/* Animated dashed flow connectors */
@keyframes gm-dash-flow {
  to { stroke-dashoffset: -24; }
}

.svg-diagram-svg path[stroke-dasharray] {
  animation: gm-dash-flow 1.2s linear infinite;
}

/* Error cards reuse the palette */
.svg-diagram-error {
  overflow: visible;
}

@media (prefers-reduced-motion: reduce) {
  .svg-diagram-svg path[stroke-dasharray] {
    animation: none;
  }
}
Use code with caution.

7. Caching & Error Handling


Content-Addressed Caching (cache.go)

• Key Formula: sha256(ModuleVersion + ":" + lang + ":" + theme + ":" + attrs + ":" + source)
• ModuleVersion salts the key; bump it whenever renderer output changes so stale fragments cannot survive an upgrade.
• Built-in Stores:
	• NewMemoryLRU(maxEntries int) — Thread-safe in-memory LRU cache (default 512 entries).
	• NewDiskCache(dir string) — Atomic file writes (CreateTemp + Rename) storing <hash>.svg on disk so compiled diagrams survive Caddy restarts.

Never-Fail Error Modes (errors.go)

• ErrorSVGCard (Default): Catches parse/layout/timeout errors and emits a valid <svg viewBox="0 0 600 120" class="svg-diagram-svg svg-diagram-error"> displaying the fence language, line number (when available), and error message, painted with --diagram-error-* and --diagram-* variables.
• ErrorCodeBlock: Falls back to standard <pre><code class="language-...">.
• ErrorFail: Returns the Go error from markdown.Convert(). It is the only mode that fails, and it is opt-in.
• Engine errors are wrapped with a diagrams: prefix so hosts can distinguish them.

8. Public Go API (diagrams.go)

The root package is a facade over pkg/render and pkg/cache; keep it free of logic. Types are aliases (ErrorMode, Cache, Config, Option).

go
package diagrams

import (
	"context"
	"time"

	"github.com/yuin/goldmark"
)

const ModuleVersion = "v0.7.0"

type ErrorMode string

const (
	ErrorSVGCard   ErrorMode = "svg_card"
	ErrorCodeBlock ErrorMode = "code_block"
	ErrorFail      ErrorMode = "fail"
)

type Cache interface {
	Get(key string) (svg string, ok bool)
	Set(key string, svg string)
}

type Option func(*Config)

func WithTheme(theme string) Option          // "auto" | "light" | "dark" | "neutral"
func WithCache(c Cache) Option               // Default: NewMemoryLRU(512)
func WithErrorMode(mode ErrorMode) Option    // Default: ErrorSVGCard
func WithLanguages(langs ...string) Option   // Default: all supported languages
func WithTimeout(d time.Duration) Option     // Default: 5 * time.Second
func WithD2Layout(layout string) Option      // "dagre" only; anything else is rejected
func WithID(id string) Option                // Sets data-id on <figure> for standalone renders

// New returns a goldmark.Extender for goldmark.WithExtensions(...)
func New(opts ...Option) goldmark.Extender

// RenderStandalone compiles a single diagram directly to normalized <figure><svg> HTML.
func RenderStandalone(ctx context.Context, lang, source string, opts ...Option) (string, error)

// NormalizeSVG rewrites a raw engine SVG into the single <svg> output contract.
func NormalizeSVG(rawSVG string, meta engines.Meta) (string, error)

// DefaultCSS returns the reference pure-CSS interactivity & theme stylesheet.
func DefaultCSS() string

// SupportedLanguages returns every accepted fence language.
func SupportedLanguages() []string

// Cache constructors.
func NewMemoryLRU(maxEntries int) Cache
func NewDiskCache(dir string) Cache
Use code with caution.

9. Repository File Layout

text
goldmark-diagrams/
├── diagrams.go              # Public API facade: options, New(), RenderStandalone(), NormalizeSVG(), DefaultCSS()
├── assets/diagrams.css      # Embedded reference CSS (//go:embed)
├── e2e.sh                   # Bash entry point: fixtures through the built container
├── Dockerfile               # Shared caddy-* build template for the container image
├── cmd/goldmark-diagrams/   # Container entry point: Markdown on stdin, HTML on stdout
├── docs/                    # Per-type guides and the contract reference
├── testdata/
│   ├── input/               # Fenced source fixtures, one per type
│   └── golden/              # Byte-for-byte expected HTML
└── pkg/
    ├── render/              # Config/Option, NormalizeSVG, Fragment, error cards
    ├── extender/            # goldmark.Extender, ASTTransformer, HTML node renderer
    ├── cache/               # SHA-256 key, MemoryLRU and DiskCache
    ├── engines/
    │   ├── router.go        # Dispatches lang -> pele-go, d2, or go-analyze/charts
    │   ├── pele.go          # pele-go SVG adapter
    │   ├── d2.go            # oss.terrastruct.com/d2 SVG adapter
    │   ├── charts.go        # go-analyze/charts SVG adapter
    │   └── transpilers/
    │       ├── plantuml.go  # Pure-Go PlantUML -> D2 / pele-go transpiler
    │       └── dot.go       # Pure-Go Graphviz DOT -> D2 transpiler
    ├── tests/               # End-to-end API and golden tests
    └── docscheck/           # Documentation coverage and render tests
Use code with caution.

10. Delivery Milestones

1. Milestone 1 — Core Pipeline & Engines: Delivered. diagrams.go, extender.go, normalizer.go, cache.go, errors.go, and the pele.go, d2.go, and charts.go adapters, verified by golden SVG fixtures.
2. Milestone 2 — PlantUML & DOT Transpilers: Delivered. @startuml, @startgantt/@startmindmap, and digraph fenced blocks compile through D2/pele-go.
