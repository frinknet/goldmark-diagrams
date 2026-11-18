Engineering Specification: goldmark-diagrams

Here is the complete specification for goldmark-diagrams incorporating the 3-engine pure-Go stack (pele-go/all + d2 + go-analyze/charts), zero WASM/CGo guarantees, unified SVG normalization, pure-CSS interactivity, and data-* hooks for database feedback loops.


1. Purpose & Core Constraints

goldmark-diagrams is a 100% pure-Go Goldmark extension (goldmark.Extender) designed for embedding inside custom Caddy builds, documentation engines, and Go web applications. It compiles fenced diagram and chart code blocks into one standardized, responsive, CSS-interactive <svg> output with optional data-* hooks for server-backed interactivity (such as Kanban card moves or chart drill-downs).

Hard Engineering Rules

1. Pure Go Only (CGO_ENABLED=0): Zero CGo, zero WASM/wazero runtimes, zero external CLI binaries (mmdc, plantuml.jar, dot), and zero headless browsers or Node/Java runtimes.
2. Single <svg> Output Contract: Every supported fenced block renders into a normalized <figure class="gm-diagram"> wrapping an inline <svg viewBox="...">. Never emit <canvas>, <img>, or inline <script> tags.
3. Never-Fail Markdown Conversion: Invalid diagram syntax or layout timeouts must never break markdown.Convert(). Errors render into an inline styled <svg> error card by default.
4. CSS-First Interactivity + Server-SVG Feedback: SVGs emit predictable data-* attributes and CSS custom properties (--gm-*) so hover dimming, animated flow lines, and dark/light themes work in pure CSS, while thin client-side listeners can round-trip mutations back to SQLite and swap the refreshed SVG.

2. Go Environment & Dependencies

• Go Version: Go 1.26+ (required by pele-go)
• Allowed Dependencies:
	• ://github.com — Markdown AST & HTML rendering
	• ://github.com (importing _ "://github.com") — Pure-Go Mermaid engine supporting all 33 Mermaid diagram types (Kanban, Mindmap, Gantt, Venn, GitGraph, Flowchart, Sequence, ERD, Class, State, C4, Sankey, Radar, Timeline, Quadrant, etc.)
	• ://terrastruct.com — Pure-Go D2 compiler & layout engine (dagre / elk)
	• ://github.com — Pure-Go SVG business charting library (charts.SVGOutput())
	• gopkg.in/yaml.v3 — Optional YAML parsing for chart specs
	• Go Standard Library (encoding/xml, crypto/sha256, regexp, sync, time)

3. Fenced Code Routing Matrix

goldmark-diagrams registers a Goldmark ASTTransformer (priority 100) that intercepts ast.FencedCodeBlock nodes, parses the info string (lang + optional key-value attributes like ```mermaid id=brd_123 theme=dark), and routes the source to one of three pure-Go engines:

Fence Language (```lang)	Parser / Preprocessor	Pure-Go Rendering Engine	Covered Diagram & Chart Types
mermaid, mmd	Direct pass-through	pele-go/all	All 33 Mermaid types: kanban, mindmap, gantt, venn-beta, gitGraph, flowchart, sequenceDiagram, erDiagram, classDiagram, stateDiagram, pie, xychart-beta, quadrantChart, sankey-beta, radar, timeline, c4, journey, etc.
d2	Direct pass-through	://terrastruct.com	Architecture, Flow, Sequence, ERD (sql_table), Class, State, Grid, Cloud Icons.
plantuml, puml	transpilers.PlantUMLToD2()	://terrastruct.com (or pele-go for @startgantt/@startmindmap)	@startuml Sequence, Class, Component, State, UseCase, Gantt, Mindmap.
dot, graphviz	transpilers.DotToD2()	://terrastruct.com	digraph, graph, subgraph cluster_*, rankdir, shapes, and edge labels.
chart, echarts	JSON / YAML unmarshal	go-analyze/charts (SVGOutput())	Line, Bar, Stacked Bar, Horizontal Bar, Pie, Donut, Radar, Funnel, Scatter, Candlestick, Heatmap, Table.


4. Standardized HTML & SVG Output Contract (normalizer.go)

Every raw <svg> produced by pele-go, d2, or go-analyze/charts must pass through NormalizeSVG(rawSVG string, meta DiagramMeta) (string, error) before caching and HTML emission.

Output Structure

html
<figure class="gm-diagram"
        data-lang="mermaid"
        data-engine="pele"
        data-theme="auto"
        data-hash="9f86d081"
        data-id="brd_123">
  <svg viewBox="0 0 900 480" class="gm-diagram-svg" role="img" aria-label="mermaid diagram">
    <!-- Normalized SVG nodes stamped with data-item, data-series, data-column-id, data-card-id, and CSS variables -->
  </svg>
</figure>
Use code with caution.

Normalization Rules

1. Responsive viewBox: Ensure viewBox="0 0 W H" exists on <svg> (deriving from width/height attributes if absent) and remove hardcoded pixel width/height attributes so CSS governs sizing.
2. Interactive data-* Stamping:
	• Propagate fence attributes (such as id=brd_123 or resource=orders) onto <figure data-id="...">.
	• Stamp data-item="<id-or-index>" onto primary visual elements (nodes, bars, pie slices, git commits, Gantt tasks).
	• Stamp data-series="<0..N>" onto multi-series groups and paths.
	• For Kanban diagrams, stamp data-column-id="<slug>" on column groups and data-card-id="<id>" on card groups so client-side pointer listeners can round-trip mutations back to the server.
	• Ensure interactive elements include a <title> child for zero-JS native hover tooltips.
3. CSS Custom Property Mapping: Map engine theme fills and strokes to CSS variables with sensible fallbacks:
	• --gm-bg (default transparent)
	• --gm-fg (default #1e293b)
	• --gm-border (default #cbd5e1)
	• --gm-series-0 through --gm-series-7

5. Server-SVG Feedback Loop (For Interactive Kanban, Gantt & Charts)

To keep state in SQLite rather than complex client-side JavaScript stores:
1. Initial Render: Your backend renders Markdown (or calls diagrams.RenderStandalone(ctx, "mermaid", kanbanSyntax, diagrams.WithID("brd_123"))) to produce the <figure data-id="brd_123"> SVG.
2. Client Interaction: A thin (~30-line) delegated pointer/click listener reads data-id, data-card-id, data-column-id, or data-item from the clicked/dropped SVG element and POSTs JSON to your backend endpoint (gated by your Caddy acl middleware).
3. SVG Swap: The backend updates SQLite, re-renders the updated diagram SVG fragment, and returns it so the client swaps figure.outerHTML = updatedHTML with zero client state drift.

6. Reference Companion CSS (diagrams.DefaultCSS())

Embeddable stylesheet shipped with the package for out-of-the-box pure-CSS interactivity:
css
.gm-diagram {
  margin: 1.5rem 0;
  overflow-x: auto;
}

.gm-diagram-svg {
  display: block;
  width: 100%;
  max-width: 100%;
  height: auto;
}

/* Pure-CSS sibling dimming on hover */
.gm-diagram-svg:has([data-item]:hover) [data-item]:not(:hover) {
  opacity: 0.35;
  transition: opacity 150ms ease;
}

.gm-diagram-svg [data-item]:hover,
.gm-diagram-svg [data-card-id]:hover {
  filter: brightness(1.08);
  cursor: pointer;
}

/* Animated dashed flow connectors */
@keyframes gm-dash-flow {
  to { stroke-dashoffset: -24; }
}

.gm-diagram-svg path[stroke-dasharray] {
  animation: gm-dash-flow 1.2s linear infinite;
}
Use code with caution.

7. Caching & Error Handling


Content-Addressed Caching (cache.go)

• Key Formula: sha256(moduleVersion + ":" + lang + ":" + theme + ":" + attrs + ":" + source)
• Built-in Stores:
	• NewMemoryLRU(maxEntries int) — Thread-safe in-memory LRU cache (default 512 entries).
	• NewDiskCache(dir string) — Atomic file writes (CreateTemp + Rename) storing <hash>.svg on disk so compiled diagrams survive Caddy restarts.

Never-Fail Error Modes (errors.go)

• ErrorSVGCard (Default): Catches parse/layout/timeout errors and emits a valid <svg viewBox="0 0 600 120" class="gm-diagram-svg gm-diagram-error"> displaying the fence language, line number (when available), and error message.
• ErrorCodeBlock: Falls back to standard <pre><code class="language-...">.
• ErrorFail: Returns the Go error from markdown.Convert().

8. Public Go API (diagrams.go)

go
package diagrams

import (
	"context"
	"time"

	"://github.com"
)

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
func WithD2Layout(layout string) Option      // "dagre" (default) | "elk"
func WithID(id string) Option                // Sets data-id on <figure> for standalone renders

// New returns a goldmark.Extender for goldmark.WithExtensions(...)
func New(opts ...Option) goldmark.Extender

// RenderStandalone compiles a single diagram directly to normalized <figure><svg> HTML
// (used by backend endpoints re-rendering interactive Kanban/Gantt SVGs after DB mutations)
func RenderStandalone(ctx context.Context, lang, source string, opts ...Option) (string, error)

// DefaultCSS returns the reference pure-CSS interactivity & theme stylesheet
func DefaultCSS() string
Use code with caution.

9. Repository File Layout

text
goldmark-diagrams/
├── go.mod                   # go 1.26, goldmark, pele-go, d2, go-analyze/charts, yaml.v3
├── diagrams.go              # Public options, New(), RenderStandalone(), DefaultCSS()
├── extender.go              # goldmark.Extender, ASTTransformer, HTML NodeRenderer
├── normalizer.go            # Responsive viewBox, data-* attribute stamping, CSS var mapping
├── cache.go                 # SHA-256 MemoryLRU and DiskCache implementations
├── errors.go                # Styled <svg> error card generator
├── assets/
│   └── diagrams.css         # Embedded reference CSS (//go:embed)
└── engines/
    ├── router.go            # Dispatches lang -> pele-go, d2, or go-analyze/charts
    ├── pele.go              # _ "://github.com" SVG adapter
    ├── d2.go                # ://terrastruct.com SVG adapter
    ├── charts.go            # ://github.com SVGOutput() adapter
    └── transpilers/
        ├── plantuml.go      # Pure-Go PlantUML -> D2 / pele-go transpiler
        └── dot.go           # Pure-Go Graphviz DOT -> D2 transpiler
Use code with caution.

10. Delivery Milestones for Your Developer

1. Milestone 1 — Core Pipeline & Engines: Implement diagrams.go, extender.go, normalizer.go, cache.go, errors.go, and wire up pele.go, d2.go, and charts.go. Verify CGO_ENABLED=0 go test ./... passes with golden SVG tests across Kanban, Mindmap, Venn, Gantt, GitGraph, D2 ERD, and ECharts Bar/Pie blocks.
2. Milestone 2 — PlantUML & DOT Transpilers: Implement plantuml.go and dot.go so @startuml and digraph fenced blocks compile cleanly through D2/pele-go.
3. Milestone 3 — Caddy Integration: Register diagrams.New(...) inside your custom Caddy templates module ({{ markdown .Content }} or {{ markdown_diagrams .Content }}) and expose a /_diagrams/css route serving diagrams.DefaultCSS().
4. Milestone 4 — Interactive DB Feedback Demo: Connect RenderStandalone() to a POST /_diagrams/kanban/move handler backed by SQLite + Caddy acl to verify drag-and-drop SVG swapping.
Would you like to pivot back to your payment + acl Caddy plugin (locking down the canonical JSON event payloads and Go module skeleton now that your ERD and goldmark-diagrams spec are complete)?

