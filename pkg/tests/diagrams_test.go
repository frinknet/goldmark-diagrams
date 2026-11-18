package tests

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	diagrams "github.com/goldmark-diagrams/goldmark-diagrams"
	"github.com/goldmark-diagrams/goldmark-diagrams/pkg/engines"
	"github.com/yuin/goldmark"
)

// repoRoot is the checkout root, three directories above this file, so the
// golden fixtures are found whatever the working directory is.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test source file")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func meta(lang string, w, h float64) engines.Meta {
	return engines.Meta{Lang: lang, Theme: "auto", Width: w, Height: h}
}

func mustMarkdown(t *testing.T, opts ...diagrams.Option) goldmark.Markdown {
	t.Helper()
	return goldmark.New(goldmark.WithExtensions(diagrams.New(opts...)))
}

func convert(t *testing.T, md goldmark.Markdown, src string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	err := md.Convert([]byte(src), &buf)
	return buf.String(), err
}

func fence(lang, body string) string {
	return "```" + lang + "\n" + body + "\n```\n"
}

func assertCleanOutput(t *testing.T, out string) {
	t.Helper()
	for _, forbidden := range []string{"<canvas", "<img", "<script"} {
		if strings.Contains(out, forbidden) {
			t.Errorf("output contains forbidden element %q", forbidden)
		}
	}
}

func TestNewReturnsExtender(t *testing.T) {
	if diagrams.New() == nil {
		t.Fatal("diagrams.New() returned nil")
	}
	if len(diagrams.SupportedLanguages()) == 0 {
		t.Fatal("diagrams.SupportedLanguages() is empty")
	}
}

func TestDefaultCSSIsEmbedded(t *testing.T) {
	css := diagrams.DefaultCSS()
	for _, want := range []string{".svg-diagram", ".svg-diagram-svg", "gm-dash-flow", "[data-card-id]"} {
		if !strings.Contains(css, want) {
			t.Errorf("diagrams.DefaultCSS() missing %q", want)
		}
	}
}

func TestConvertMermaidFlowchart(t *testing.T) {
	md := mustMarkdown(t)
	out, err := convert(t, md, fence("mermaid", "flowchart TD\n  A[Start] --> B[End]"))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	assertCleanOutput(t, out)
	for _, want := range []string{
		`<figure class="svg-diagram"`,
		`data-lang="mermaid"`,
		`data-engine="pele"`,
		`data-theme="auto"`,
		`data-hash="`,
		`<svg`,
		`viewBox="`,
		`class="`,
		`svg-diagram-svg`,
		`role="img"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestConvertLeavesOtherLanguagesAlone(t *testing.T) {
	md := mustMarkdown(t)
	out, err := convert(t, md, fence("go", "func main() {}"))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if strings.Contains(out, "svg-diagram") {
		t.Errorf("non-diagram fence was intercepted:\n%s", out)
	}
	if !strings.Contains(out, `language-go`) {
		t.Errorf("expected a plain go code block, got:\n%s", out)
	}
}

func TestWithLanguagesRestrictsInterception(t *testing.T) {
	md := mustMarkdown(t, diagrams.WithLanguages("d2"))
	out, err := convert(t, md, fence("mermaid", "flowchart TD\n A --> B"))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if strings.Contains(out, "svg-diagram") {
		t.Errorf("mermaid fence should be left alone when only d2 is enabled:\n%s", out)
	}
}

func TestNeverFailOnInvalidSyntax(t *testing.T) {
	cases := map[string]string{
		"mermaid":   fence("mermaid", "flowchart TD\n  A[["),
		"d2":        fence("d2", "a -> -> b"),
		"chart":     fence("chart", `{"type": "nope"}`),
		"chartjson": fence("chart", `{not json`),
		"dot":       fence("dot", `digraph { "unterminated`),
		"plantuml":  fence("plantuml", "@startuml\n@enduml\n"),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			md := mustMarkdown(t)
			out, err := convert(t, md, src)
			if err != nil {
				t.Fatalf("Convert must never fail: %v", err)
			}
			if !strings.Contains(out, "svg-diagram-error") {
				t.Errorf("expected an error card, got:\n%s", out)
			}
			assertCleanOutput(t, out)
			if !strings.Contains(out, "diagram failed") {
				t.Errorf("error card missing message:\n%s", out)
			}
		})
	}
}

func TestErrorModeCodeBlock(t *testing.T) {
	md := mustMarkdown(t, diagrams.WithErrorMode(diagrams.ErrorCodeBlock))
	out, err := convert(t, md, fence("mermaid", "flowchart TD\n  A[["))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if !strings.Contains(out, `<pre><code class="language-mermaid">`) {
		t.Errorf("expected a code block fallback, got:\n%s", out)
	}
}

func TestErrorModeFail(t *testing.T) {
	md := mustMarkdown(t, diagrams.WithErrorMode(diagrams.ErrorFail))
	_, err := convert(t, md, fence("mermaid", "flowchart TD\n  A[["))
	if err == nil {
		t.Fatal("expected Convert to return the engine error")
	}
	if !strings.Contains(err.Error(), "diagrams:") {
		t.Errorf("error should be wrapped, got: %v", err)
	}
}

func TestInfoAttributesStampTheFigure(t *testing.T) {
	md := mustMarkdown(t)
	out, err := convert(t, md, "```mermaid id=brd_123 theme=dark\nkanban\n  Todo\n    [Task one]\n```\n")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	for _, want := range []string{`data-id="brd_123"`, `data-theme="dark"`} {
		if !strings.Contains(out, want) {
			t.Errorf("figure missing %q\n%s", want, out)
		}
	}
	// The fence theme is the figure theme, not a second data-theme attribute.
	if n := strings.Count(out, "data-theme="); n != 1 {
		t.Errorf("expected exactly one data-theme attribute, got %d:\n%s", n, out)
	}
}

func TestKanbanStampsColumnsAndCards(t *testing.T) {
	md := mustMarkdown(t)
	out, err := convert(t, md, fence("mermaid", "kanban\n  Todo\n    [Task one]\n  Done\n"))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	for _, want := range []string{
		`data-column-id="todo"`,
		`data-role="column"`,
		`data-card-id="Task one"`,
		`data-role="card"`,
		`<title>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("kanban output missing %q\n%s", want, out)
		}
	}
	// A card must not also be a data-item, or hovering it would dim the column
	// that contains it (the reference CSS dims non-hovered [data-item]s).
	if strings.Contains(out, `data-item="Task one"`) {
		t.Errorf("cards must not carry data-item:\n%s", out)
	}
}

func TestEveryTypeCarriesTheSharedHooks(t *testing.T) {
	// Every supported engine must produce identifiable elements, not just a
	// wrapper, so a client-side handler can attach to each mark.
	cases := []struct {
		name    string
		lang    string
		source  string
		wantAll []string
	}{
		{
			name:    "mermaid",
			lang:    "mermaid",
			source:  "flowchart TD\n  A[Start] --> B[End]",
			wantAll: []string{`data-item="A"`, `data-item="B"`, `data-role="node"`, `data-role="edge"`, `svg-node`, `svg-edge`},
		},
		{
			name:    "d2",
			lang:    "d2",
			source:  "alpha -> beta",
			wantAll: []string{`data-item="alpha"`, `data-item="beta"`, `data-role="shape"`, `data-role="connection"`, `svg-shape`, `svg-connection`},
		},
		{
			name:    "dot",
			lang:    "dot",
			source:  "digraph { a -> b }",
			wantAll: []string{`data-role="shape"`, `data-role="connection"`, `svg-shape`, `svg-connection`},
		},
		{
			name:    "plantuml",
			lang:    "plantuml",
			source:  "@startuml\nA -> B: hi\n@enduml",
			wantAll: []string{`data-role="shape"`, `svg-shape`},
		},
		{
			name:    "chart",
			lang:    "chart",
			source:  `{"type":"bar","labels":["Mon","Tue"],"series":[[1,2]]}`,
			wantAll: []string{`data-item="Mon"`, `data-item="Tue"`, `data-series="0"`, `data-role="mark"`, `svg-mark`, `svg-series-0`},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			md := mustMarkdown(t)
			out, err := convert(t, md, fence(tc.lang, tc.source))
			if err != nil {
				t.Fatalf("Convert: %v", err)
			}
			assertCleanOutput(t, out)
			if strings.Contains(out, "svg-diagram-error") {
				t.Fatalf("unexpected error card:\n%s", out)
			}
			for _, want := range tc.wantAll {
				if !strings.Contains(out, want) {
					t.Errorf("missing %q\n%s", want, out)
				}
			}
		})
	}
}

func TestChartsStampSeries(t *testing.T) {
	md := mustMarkdown(t)
	spec := `{"type":"bar","labels":["a","b"],"series":[[1,2],[3,4]]}`
	out, err := convert(t, md, fence("chart", spec))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if !strings.Contains(out, `data-series="0"`) {
		t.Errorf("chart output missing data-series:\n%s", out)
	}
	if !strings.Contains(out, `data-item="a"`) || !strings.Contains(out, `data-item="b"`) {
		t.Errorf("chart marks are not named by category:\n%s", out)
	}
	if !strings.Contains(out, `var(--diagram-series-0,`) {
		t.Errorf("chart fills are not mapped to --diagram-series-*:\n%s", out)
	}
	if !strings.Contains(out, `svg-mark`) || !strings.Contains(out, `svg-series-0`) {
		t.Errorf("chart marks are not styled by a standard class:\n%s", out)
	}
}

func TestD2ConnectionsAreNamedByEndpoints(t *testing.T) {
	md := mustMarkdown(t)
	source := "alpha -> beta\nalpha -> beta: again\nbeta -> gamma\ngamma -> beta"
	out, err := convert(t, md, fence("d2", source))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	assertCleanOutput(t, out)
	for _, want := range []string{
		`data-item="alpha-&gt;beta"`,
		`data-item="alpha-&gt;beta[1]"`,
		`data-item="beta-&gt;gamma"`,
		`data-item="gamma-&gt;beta"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("connection hook missing %q\n%s", want, out)
		}
	}
	for _, a := range regexp.MustCompile(`data-item="([^"]*)" data-role="connection"`).FindAllStringSubmatch(out, -1) {
		if _, err := strconv.Atoi(a[1]); err == nil {
			t.Errorf("connection %q is still named by a counter", a[1])
		}
	}
}

func TestChartsNameUnlabelledMarksByCategory(t *testing.T) {
	md := mustMarkdown(t)
	out, err := convert(t, md, fence("chart", `{"type":"bar","series":[[1,2,3],[4,5,6]]}`))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	for _, want := range []string{
		`data-item="1" data-series="0" data-role="mark"`,
		`data-item="2" data-series="0" data-role="mark"`,
		`data-item="3" data-series="0" data-role="mark"`,
		`data-item="1" data-series="1" data-role="mark"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("mark hook missing %q\n%s", want, out)
		}
	}
}

func TestD2PaletteIsMappedToSharedVariables(t *testing.T) {
	md := mustMarkdown(t)
	out, err := convert(t, md, fence("d2", "a -> b"))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	for _, want := range []string{
		"var(--diagram-bg,",
		"var(--diagram-fg,",
		"var(--diagram-border,",
		"var(--diagram-accent,",
		"var(--diagram-surface,",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("d2 palette not mapped: missing %q", want)
		}
	}
	if strings.Contains(out, ".fill-N7{fill:#") {
		t.Errorf("d2 neutral ramp was not rewritten:\n%s", out)
	}
}

func TestPeleVariablesAliasTheSharedPalette(t *testing.T) {
	md := mustMarkdown(t)
	out, err := convert(t, md, fence("mermaid", "flowchart TD\n  A --> B"))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if !strings.Contains(out, "--pele-fg:var(--diagram-fg,") {
		t.Errorf("pele variables are not aliased onto --diagram-*:\n%s", out)
	}
}

func TestDefaultCSSDefinesThePalette(t *testing.T) {
	css := diagrams.DefaultCSS()
	for _, want := range []string{
		"--diagram-fg:", "--diagram-border:", "--diagram-surface:",
		"--diagram-series-0:", "--diagram-series-7:",
		"prefers-color-scheme: dark",
		`[data-theme="dark"]`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("diagrams.DefaultCSS() missing %q", want)
		}
	}
}

func TestRenderStandaloneWithID(t *testing.T) {
	out, err := diagrams.RenderStandalone(context.Background(), "mermaid", "flowchart TD\n  A --> B", diagrams.WithID("brd_123"))
	if err != nil {
		t.Fatalf("RenderStandalone: %v", err)
	}
	if !strings.Contains(out, `data-id="brd_123"`) {
		t.Errorf("standalone render missing data-id:\n%s", out)
	}
}

func TestRenderStandaloneRespectsErrorMode(t *testing.T) {
	_, err := diagrams.RenderStandalone(context.Background(), "mermaid", "flowchart TD\n  A[[", diagrams.WithErrorMode(diagrams.ErrorFail))
	if err == nil {
		t.Fatal("expected an error with diagrams.ErrorFail")
	}
	if _, err := diagrams.RenderStandalone(context.Background(), "mermaid", "flowchart TD\n  A[["); err != nil {
		t.Fatalf("default mode must not fail: %v", err)
	}
}

func TestRenderStandaloneWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out, err := diagrams.RenderStandalone(ctx, "mermaid", "flowchart TD\n  A --> B")
	if err != nil {
		t.Fatalf("canceled render must not fail markdown: %v", err)
	}
	if !strings.Contains(out, "svg-diagram-error") {
		t.Errorf("expected an error card for a canceled render:\n%s", out)
	}
}

func TestNormalizeSVGSynthesizesViewBox(t *testing.T) {
	got, err := diagrams.NormalizeSVG(`<?xml version="1.0"?><svg width="320px" height="200" xmlns="http://www.w3.org/2000/svg"><rect/></svg>`, meta("d2", 0, 0))
	if err != nil {
		t.Fatalf("NormalizeSVG: %v", err)
	}
	if !strings.Contains(got, `viewBox="0 0 320 200"`) {
		t.Errorf("viewBox not synthesized:\n%s", got)
	}
	if strings.Contains(got, `width="320px"`) || strings.Contains(got, ` height="200"`) {
		t.Errorf("pixel size not stripped:\n%s", got)
	}
	if strings.HasPrefix(got, "<?xml") {
		t.Errorf("XML prolog not stripped:\n%s", got)
	}
}

func TestNormalizeSVGRejectsNonSVG(t *testing.T) {
	if _, err := diagrams.NormalizeSVG("not a diagram", meta("d2", 0, 0)); err == nil {
		t.Fatal("expected an error for input with no <svg>")
	}
}

func TestNormalizeSVGLeavesStyleContentAlone(t *testing.T) {
	got, err := diagrams.NormalizeSVG(`<svg viewBox="0 0 10 10"><style>a{b:c}</style></svg>`, meta("d2", 10, 10))
	if err != nil {
		t.Fatalf("NormalizeSVG: %v", err)
	}
	if !strings.Contains(got, "<style>a{b:c}</style>") {
		t.Errorf("style content was mangled:\n%s", got)
	}
}

func TestCacheIsConsulted(t *testing.T) {
	cache := &countingCache{data: map[string]string{}}
	md := mustMarkdown(t, diagrams.WithCache(cache))
	src := fence("mermaid", "flowchart TD\n  A --> B")
	for i := 0; i < 2; i++ {
		if _, err := convert(t, md, src); err != nil {
			t.Fatalf("Convert: %v", err)
		}
	}
	if cache.sets != 1 {
		t.Errorf("expected 1 cache set, got %d", cache.sets)
	}
	if cache.hits != 1 {
		t.Errorf("expected 1 cache hit, got %d", cache.hits)
	}
}

func TestMemoryLRUEvicts(t *testing.T) {
	c := diagrams.NewMemoryLRU(2)
	c.Set("a", "1")
	c.Set("b", "2")
	c.Set("c", "3")
	if _, ok := c.Get("a"); ok {
		t.Error("expected the least recently used entry to be evicted")
	}
	if v, ok := c.Get("c"); !ok || v != "3" {
		t.Errorf("Get(c) = %q, %v", v, ok)
	}
	c.Set("b", "20")
	if v, _ := c.Get("b"); v != "20" {
		t.Errorf("Set should overwrite, got %q", v)
	}
}

func TestDiskCacheRoundTrips(t *testing.T) {
	dir := t.TempDir()
	c := diagrams.NewDiskCache(dir)
	if _, ok := c.Get("missing"); ok {
		t.Error("expected a miss on an empty cache")
	}
	c.Set("k", "<figure></figure>")
	if v, ok := c.Get("k"); !ok || v != "<figure></figure>" {
		t.Errorf("Get(k) = %q, %v", v, ok)
	}
	if _, err := os.Stat(filepath.Join(dir, "k.svg")); err != nil {
		t.Errorf("expected k.svg on disk: %v", err)
	}
}

func TestGolden(t *testing.T) {
	update := os.Getenv("UPDATE_GOLDEN") == "1"
	root := repoRoot(t)
	inputs, err := filepath.Glob(filepath.Join(root, "testdata", "input", "*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 {
		t.Fatal("no fixtures in testdata/input")
	}
	sort.Strings(inputs)
	for _, path := range inputs {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		lang := strings.TrimPrefix(filepath.Ext(path), ".")
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			md := mustMarkdown(t)
			out, err := convert(t, md, fence(lang, strings.TrimRight(string(source), "\n")))
			if err != nil {
				t.Fatalf("Convert: %v", err)
			}
			golden := filepath.Join(root, "testdata", "golden", name+".html")
			if update {
				if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(golden, []byte(out), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("missing golden file (run UPDATE_GOLDEN=1 go test ./...): %v", err)
			}
			if out != string(want) {
				t.Errorf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, out, want)
			}
		})
	}
}

type countingCache struct {
	mu   sync.Mutex
	data map[string]string
	gets int
	sets int
	hits int
}

func (c *countingCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gets++
	if v, ok := c.data[key]; ok {
		c.hits++
		return v, true
	}
	return "", false
}

func (c *countingCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sets++
	c.data[key] = value
}
