// Package docscheck keeps the reference documentation honest: every diagram
// type the engines accept has a page, every page is well formed, and every
// example on a page still renders.
package docscheck

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	diagrams "github.com/goldmark-diagrams/goldmark-diagrams"
)

// Every type in the Mermaid table of spec.md and every chart type the router
// accepts must have a page, so the reference cannot silently fall behind the
// engines.
var documentedMermaidTypes = []string{
	"agentflow.md", "architecture.md", "block.md", "c4.md", "class.md",
	"cynefin.md", "er.md", "eventmodeling.md", "flowchart.md", "gantt.md",
	"git.md", "info.md", "ishikawa.md", "journey.md", "kanban.md",
	"mindmap.md", "packet.md", "pie.md", "quadrant.md", "radar.md",
	"railroad.md", "requirement.md", "sankey.md", "sequence.md", "state.md",
	"swimlane.md", "timeline.md", "treemap.md", "treeview.md", "usecase.md",
	"venn.md", "wardley.md", "xychart.md",
}

var documentedChartTypes = []string{
	"bar.md", "candlestick.md", "donut.md", "funnel.md", "heatmap.md",
	"horizontal-bar.md", "line.md", "pie.md", "radar.md", "scatter.md",
	"stacked-bar.md", "table.md",
}

var docsLinkRE = regexp.MustCompile(`\]\(([^()#]+\.md)\)`)

// repoRoot is the checkout root, three directories above this file, so the
// tests can reach docs/ whatever the working directory is.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test source file")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func TestDocsCoverEveryDiagramType(t *testing.T) {
	root := repoRoot(t)
	cases := []struct {
		dir  string
		want []string
	}{
		{filepath.Join(root, "docs", "mermaid"), documentedMermaidTypes},
		{filepath.Join(root, "docs", "charts"), documentedChartTypes},
	}
	for _, tc := range cases {
		entries, err := os.ReadDir(tc.dir)
		if err != nil {
			t.Fatalf("ReadDir %s: %v", tc.dir, err)
		}
		got := map[string]bool{}
		for _, e := range entries {
			if !e.IsDir() {
				got[e.Name()] = true
			}
		}
		for _, name := range tc.want {
			if !got[name] {
				t.Errorf("%s/%s is missing", tc.dir, name)
			}
			delete(got, name)
		}
		for name := range got {
			t.Errorf("%s/%s documents no known type", tc.dir, name)
		}
	}
}

func TestDocsPagesAreWellFormed(t *testing.T) {
	root := repoRoot(t)
	docsDir := filepath.Join(root, "docs")
	indexPath := filepath.Join(docsDir, "README.md")

	var pages []string
	err := filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			pages = append(pages, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk docs: %v", err)
	}
	if len(pages) == 0 {
		t.Fatal("no documentation pages found")
	}
	for _, path := range pages {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile %s: %v", path, err)
		}
		page := string(b)
		if !strings.HasPrefix(page, "# ") {
			t.Errorf("%s does not start with an H1", path)
		}
		if path == indexPath {
			continue // the index links instead of showing examples
		}
		if !strings.Contains(page, "```") {
			t.Errorf("%s has no fenced example", path)
		}
	}

	index, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", indexPath, err)
	}
	for _, m := range docsLinkRE.FindAllStringSubmatch(string(index), -1) {
		target := filepath.Join(docsDir, filepath.FromSlash(m[1]))
		if _, err := os.Stat(target); err != nil {
			t.Errorf("docs/README.md links to missing %s", m[1])
		}
	}
}

// extractFence returns the body of the first ```<tag> block in a page.
func extractFence(page, tag string) (string, bool) {
	lines := strings.Split(page, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "```"+tag {
			continue
		}
		var body []string
		for _, bodyLine := range lines[i+1:] {
			if strings.TrimSpace(bodyLine) == "```" {
				return strings.Join(body, "\n"), true
			}
			body = append(body, bodyLine)
		}
	}
	return "", false
}

// The example on every type page must actually render. A page whose snippet is
// not valid for the engine it documents would ship an error card to readers.
func TestDocsExamplesRender(t *testing.T) {
	root := repoRoot(t)
	cases := []struct {
		path    string
		fence   string
		lang    string
		pattern string
	}{
		{filepath.Join(root, "docs", "mermaid", "*.md"), "mermaid", "mermaid", "*.md"},
		{filepath.Join(root, "docs", "charts", "*.md"), "json", "chart", "*.md"},
		{filepath.Join(root, "docs", "languages", "d2.md"), "d2", "d2", ""},
		{filepath.Join(root, "docs", "languages", "dot.md"), "dot", "dot", ""},
		{filepath.Join(root, "docs", "languages", "plantuml.md"), "plantuml", "plantuml", ""},
	}
	for _, tc := range cases {
		var paths []string
		if tc.pattern == "" {
			paths = []string{tc.path}
		} else {
			matches, err := filepath.Glob(tc.path)
			if err != nil {
				t.Fatalf("Glob %s: %v", tc.path, err)
			}
			paths = matches
		}
		for _, path := range paths {
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile %s: %v", path, err)
			}
			src, ok := extractFence(string(b), tc.fence)
			if !ok || strings.TrimSpace(src) == "" {
				t.Errorf("%s has no ```%s example", path, tc.fence)
				continue
			}
			out, err := diagrams.RenderStandalone(context.Background(), tc.lang, src)
			if err != nil {
				t.Errorf("%s example failed to render: %v", path, err)
				continue
			}
			if strings.Contains(out, "svg-diagram-error") {
				t.Errorf("%s example renders the error card:\n%s", path, out)
			}
		}
	}
}
