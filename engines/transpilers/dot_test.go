package transpilers

import (
	"strings"
	"testing"
)

func TestDotToD2(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "simple digraph",
			src:  "digraph G {\n  a -> b;\n}\n",
			want: "a -> b",
		},
		{
			name: "undirected graph",
			src:  "graph G { a -- b }",
			want: "a -> b",
		},
		{
			name: "strict prefix",
			src:  "strict digraph G { a -> b }",
			want: "a -> b",
		},
		{
			name: "bare body",
			src:  "{ a -> b }",
			want: "a -> b",
		},
		{
			name: "rankdir LR",
			src:  "digraph G { rankdir=LR; a -> b }",
			want: "direction: right\na -> b",
		},
		{
			name: "rankdir BT",
			src:  "digraph G { rankdir = BT; a }",
			want: "direction: up\na",
		},
		{
			name: "graph attributes in brackets",
			src:  "digraph G { graph [rankdir=TB, label=\"Top\"]; a }",
			want: "direction: down\ntitle: Top\na",
		},
		{
			name: "graph label",
			src:  "digraph G { label=\"My graph\"; a }",
			want: "title: My graph\na",
		},
		{
			name: "node attributes",
			src:  "digraph G { a [label=\"Alpha\", shape=box] }",
			want: "a\na.shape: rectangle\na: Alpha",
		},
		{
			name: "bare node statement",
			src:  "digraph G { a }",
			want: "a",
		},
		{
			name: "ids with digits and underscores",
			src:  "digraph G { node_1 -> node2 }",
			want: "node_1 -> node2",
		},
		{
			name: "quoted node ids",
			src:  "digraph G { \"a b\" -> c }",
			want: "\"a b\" -> c",
		},
		{
			name: "edge label",
			src:  "digraph G { a -> b [label=\"yes\"] }",
			want: "a -> b: yes",
		},
		{
			name: "edge chain",
			src:  "digraph G { a -> b -> c }",
			want: "a -> b\nb -> c",
		},
		{
			name: "node defaults apply to later nodes",
			src:  "digraph G { node [shape=box]; a; b }",
			want: "a\na.shape: rectangle\nb\nb.shape: rectangle",
		},
		{
			name: "explicit attributes override node defaults",
			src:  "digraph G { node [shape=box]; a [shape=diamond] }",
			want: "a\na.shape: diamond",
		},
		{
			name: "node defaults apply to edge endpoints",
			src:  "digraph G { node [shape=box]; a -> b }",
			want: "a.shape: rectangle\nb.shape: rectangle\na -> b",
		},
		{
			name: "edge defaults ignored",
			src:  "digraph G { edge [color=red]; a -> b }",
			want: "a -> b",
		},
		{
			name: "cluster subgraph",
			src:  "digraph G { subgraph cluster_foo { label=\"Group\"; a; b } c -> a }",
			want: "foo: {\n  label: Group\n  a\n  b\n}\nc -> foo.a",
		},
		{
			name: "named subgraph",
			src:  "digraph G { subgraph one { a } }",
			want: "one: {\n  a\n}",
		},
		{
			name: "nested subgraphs",
			src:  "digraph G { subgraph cluster_a { subgraph cluster_b { x } } }",
			want: "a: {\n  b: {\n    x\n  }\n}",
		},
		{
			name: "comments",
			src:  "// line\n# hash\n/* block */\ndigraph G {\n  a -> b // trailing\n}\n",
			want: "a -> b",
		},
		{
			name: "ports are dropped",
			src:  "digraph G { a:p -> b:q }",
			want: "a -> b",
		},
		{
			name: "multi line label is quoted",
			src:  "digraph G { a [label=\"one\\ntwo\"] }",
			want: "a\na: \"one\\ntwo\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DotToD2(tt.src)
			if err != nil {
				t.Fatalf("DotToD2(%q) returned error: %v", tt.src, err)
			}
			if want := tt.want + "\n"; got != want {
				t.Errorf("DotToD2(%q) =\n%q\nwant\n%q", tt.src, got, want)
			}
		})
	}
}

func TestDotToD2Shapes(t *testing.T) {
	tests := []struct {
		shape string
		want  string
	}{
		{"box", "rectangle"},
		{"rect", "rectangle"},
		{"rectangle", "rectangle"},
		{"ellipse", "oval"},
		{"oval", "oval"},
		{"circle", "circle"},
		{"doublecircle", "circle"},
		{"point", "circle"},
		{"diamond", "diamond"},
		{"hexagon", "hexagon"},
		{"cylinder", "cylinder"},
		{"cylinder_shape", "cylinder"},
		{"box3d", "cylinder"},
		{"parallelogram", "parallelogram"},
		{"trapezium", "trapezoid"},
		{"house", "rectangle"},
		{"invhouse", "rectangle"},
		{"record", "rectangle"},
		{"Mrecord", "rectangle"},
		{"plaintext", "rectangle"},
		{"none", "rectangle"},
		{"cloud", "cloud"},
		{"folder", "rectangle"},
		{"note", "rectangle"},
		{"tab", "rectangle"},
		{"spikes", "rectangle"},
	}

	for _, tt := range tests {
		t.Run(tt.shape, func(t *testing.T) {
			src := "digraph G { a [shape=" + tt.shape + "] }"
			want := "a\na.shape: " + tt.want + "\n"
			got, err := DotToD2(src)
			if err != nil {
				t.Fatalf("DotToD2(%q) returned error: %v", src, err)
			}
			if got != want {
				t.Errorf("DotToD2(%q) = %q, want %q", src, got, want)
			}
		})
	}
}

func TestDotToD2SubgraphLabel(t *testing.T) {
	src := "digraph G {\n  subgraph cluster_db {\n    label=\"Database tier\";\n    db1\n  }\n}\n"
	got, err := DotToD2(src)
	if err != nil {
		t.Fatalf("DotToD2(%q) returned error: %v", src, err)
	}
	if !strings.Contains(got, "db: {") {
		t.Errorf("output missing container:\n%s", got)
	}
	// The container label is emitted as `label:` inside `db: {`, i.e. db.label.
	if !strings.Contains(got, "label: Database tier") {
		t.Errorf("output missing container label:\n%s", got)
	}
}

func TestDotToD2Errors(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr string
	}{
		{"empty source", "", "dot: no statements"},
		{"empty graph", "digraph G {}", "dot: no statements"},
		{"comment only", "// nothing here\n", "dot: no statements"},
		{"graph keyword only", "digraph", "dot: no statements"},
		{"unterminated string", "digraph G { a [label=\"oops] }", "unterminated quoted string"},
		{"unterminated comment", "digraph G { /* oops", "unterminated comment"},
		{"missing closing brace", "digraph G { a ", "missing closing brace"},
		{"missing subgraph brace", "digraph G { subgraph }", "expected '{'"},
		{"dangling edge", "digraph G { a -> }", "expected identifier"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DotToD2(tt.src)
			if err == nil {
				t.Fatalf("DotToD2(%q) = %q, want error", tt.src, got)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("DotToD2(%q) error = %q, want it to contain %q", tt.src, err.Error(), tt.wantErr)
			}
			if got != "" {
				t.Errorf("DotToD2(%q) returned %q alongside an error", tt.src, got)
			}
		})
	}
}

// TestDotToD2NeverPanics feeds malformed input to the transpiler; a panic
// fails the test.
func TestDotToD2NeverPanics(t *testing.T) {
	inputs := []string{
		"",
		" ",
		"}",
		"->",
		"a ->",
		"-> a",
		"\"",
		"[",
		"]",
		"a [",
		"a [label=]",
		"a [label",
		"a =",
		"= a",
		"digraph {",
		"digraph { subgraph",
		"subgraph",
		"subgraph {",
		"digraph G { a -> b -> }",
		"/*",
		"/",
		"#",
		"digraph G { a [shape=] }",
		"node",
		"node [",
		"edge [",
		"graph",
		"\x00\x01\x02",
		"digraph \"quoted name\" { a }",
		"digraph G { a -> b [label=\"x\" }",
		"digraph G { a:b -> c }",
	}
	for _, in := range inputs {
		_, _ = DotToD2(in)
	}
}
