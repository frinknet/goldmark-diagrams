package transpilers

import (
	"strings"
	"testing"
)

func TestPlantUMLToD2(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "sequence basics",
			src: `@startuml
title Login flow
participant Alice
actor Bob as B
Alice -> B: hello
B --> Alice : hi
activate B
deactivate B
autonumber
== phase two ==
@enduml
`,
			want: `title: Login flow
Alice
B: Bob
B.shape: person
Alice -> B: hello
B -> Alice: hi`,
		},
		{
			name: "sequence blocks are flattened into containers",
			src: `@startuml
Alice -> Bob: start
alt happy path
Alice -> Bob: next
else broken
Alice -> Bob: retry
end
loop retry
Bob -> Alice: again
end
@enduml
`,
			want: `Alice -> Bob: start
happy_path: {
  label: happy path
  Alice -> Bob: next
}
broken: {
  Alice -> Bob: retry
}
retry: {
  Bob -> Alice: again
}`,
		},
		{
			name: "class diagram",
			src: `@startuml
class User {
+name: String
+login()
}
interface Repo
abstract class Base
enum Color
User <|-- Admin
User <|.. Service
User *-- Address
Foo o-- Bar
A --> B
C ..> D
E -- F
@enduml
`,
			want: `User: {
  label: "User\n+name: String\n+login()"
}
Repo
Repo.shape: class
Base
Base.shape: class
Color
Color.shape: class
Admin -> User: inherits
Service -> User: inherits
User -> Address: composition
Foo -> Bar: aggregation
A -> B
C -> D: depends on
E -> F`,
		},
		{
			name: "component state and usecase",
			src: `@startuml
component Web
state Idle
usecase (Login)
[Order system] --> [Database]
Web -> Idle
@enduml
`,
			want: `Web
Idle
Login
Login.shape: oval
Order_system: Order system
Order_system -> Database
Web -> Idle`,
		},
		{
			name: "notes inline and block",
			src: `@startuml
Alice -> Bob: hi
note over Alice: remember this
note left of Alice
first line
second line
end note
@enduml
`,
			want: `Alice -> Bob: hi
note1: remember this
note1.shape: callout
note2: "first line\nsecond line"
note2.shape: callout`,
		},
		{
			name: "left to right direction",
			src:  "@startuml\nleft to right direction\nAlice -> Bob: hi\n@enduml\n",
			want: "direction: right\nAlice -> Bob: hi",
		},
		{
			name: "declaration keyword used as an edge source",
			src:  "@startuml\nstate --> Idle\n@enduml\n",
			want: "state -> Idle",
		},
		{
			name: "bracketed usecase nodes",
			src: `@startuml
[Login]
[Dashboard] --> [Login]
@enduml
`,
			want: `Login
Dashboard -> Login`,
		},
		{
			name: "class alias and enum members",
			src: `@startuml
class "Order Service" as OS {
+submit()
}
enum Status {
NEW
}
@enduml
`,
			want: `OS: {
  label: "Order Service\n+submit()"
}
Status: {
  label: "Status\nNEW"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PlantUMLToD2(tt.src)
			if err != nil {
				t.Fatalf("PlantUMLToD2(%q) returned error: %v", tt.src, err)
			}
			if want := tt.want + "\n"; got != want {
				t.Errorf("PlantUMLToD2(%q) =\n%q\nwant\n%q", tt.src, got, want)
			}
		})
	}
}

func TestPlantUMLToD2TitleOnly(t *testing.T) {
	got, err := PlantUMLToD2("@startuml\ntitle Solo title\n@enduml\n")
	if err != nil {
		t.Fatalf("PlantUMLToD2 returned error: %v", err)
	}
	if want := "title: Solo title\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPlantUMLToD2Errors(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr string
	}{
		{"empty source", "", "plantuml: no statements"},
		{"empty diagram", "@startuml\n@enduml\n", "plantuml: no statements"},
		{"skinparam only", "@startuml\nskinparam monochrome true\n@enduml\n", "plantuml: no statements"},
		{"prose only", "just some words", "plantuml: no statements"},
		{
			"gantt belongs to mermaid",
			"@startgantt\n[Task one] lasts 10 days\n@endgantt\n",
			"plantuml: unsupported diagram",
		},
		{
			"mindmap belongs to mermaid",
			"@startmindmap\n* root\n@endmindmap\n",
			"plantuml: unsupported diagram",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PlantUMLToD2(tt.src)
			if err == nil {
				t.Fatalf("PlantUMLToD2(%q) = %q, want error", tt.src, got)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("PlantUMLToD2(%q) error = %q, want it to contain %q", tt.src, err.Error(), tt.wantErr)
			}
			if got != "" {
				t.Errorf("PlantUMLToD2(%q) returned %q alongside an error", tt.src, got)
			}
		})
	}
}

// TestPlantUMLToD2NeverPanics feeds malformed input to the transpiler; a panic
// fails the test.
func TestPlantUMLToD2NeverPanics(t *testing.T) {
	inputs := []string{
		"",
		" ",
		"@startuml",
		"@startuml\n}",
		"@startuml\nend\nalt\nelse\n@enduml",
		"@startuml\nnote",
		"@startuml\nnote over\n@enduml",
		"@startuml\nnote over Alice\nno end note here\n@enduml",
		"@startuml\nclass",
		"@startuml\nclass A {\n@enduml",
		"@startuml\nclass A {",
		"@startuml\n<|--",
		"@startuml\n--> @enduml",
		"@startuml\n[",
		"@startuml\n[]",
		"@startuml\nparticipant",
		"@startuml\nparticipant as",
		"@startuml\nactor\n@enduml",
		"@startuml\ntitle\n@enduml",
		"@startuml\nusecase\n@enduml",
		"@startuml\n@enduml",
		"\x00\x01",
	}
	for _, in := range inputs {
		_, _ = PlantUMLToD2(in)
	}
}

func TestPlantUMLToMermaidGantt(t *testing.T) {
	src := `@startgantt
Project starts 2024-01-01
title My plan
-- Design --
[Task one] lasts 10 days
[Task two] starts after [Task one]
[Task three] starts at 2024-02-01
[Task four] is colored red
[Task five] lasts 2 weeks
@endgantt
`
	want := `gantt
  dateFormat YYYY-MM-DD
  title My plan
  section Design
  Task one :t1, 2024-01-01, 10d
  Task two :t2, after t1, 1d
  Task three :t3, 2024-02-01, 1d
  Task four :t4, 2024-01-01, 1d
  Task five :t5, 2024-01-01, 14d
`
	got, err := PlantUMLToMermaid(src)
	if err != nil {
		t.Fatalf("PlantUMLToMermaid returned error: %v", err)
	}
	if got != want {
		t.Errorf("PlantUMLToMermaid(%q) =\n%q\nwant\n%q", src, got, want)
	}
}

func TestPlantUMLToMermaidGanttWithoutProjectStart(t *testing.T) {
	src := `@startgantt
[First] lasts 3 days
[Second] starts after [First]
-- Later --
[Third] starts at 2024-03-01
@endgantt
`
	want := `gantt
  dateFormat YYYY-MM-DD
  First :3d
  Second :t1, after task1, 1d
  section Later
  Third :t2, 2024-03-01, 1d
`
	got, err := PlantUMLToMermaid(src)
	if err != nil {
		t.Fatalf("PlantUMLToMermaid returned error: %v", err)
	}
	if got != want {
		t.Errorf("PlantUMLToMermaid(%q) =\n%q\nwant\n%q", src, got, want)
	}
}

func TestPlantUMLToMermaidMindmap(t *testing.T) {
	src := `@startmindmap
* Project
** Design
*** Wireframes
** Build
-- Testing
@endmindmap
`
	want := `mindmap
  root((Project))
    Design
      Wireframes
    Build
    Testing
`
	got, err := PlantUMLToMermaid(src)
	if err != nil {
		t.Fatalf("PlantUMLToMermaid returned error: %v", err)
	}
	if got != want {
		t.Errorf("PlantUMLToMermaid(%q) =\n%q\nwant\n%q", src, got, want)
	}
}

func TestPlantUMLToMermaidMindmapExtraRootBecomesChild(t *testing.T) {
	src := "@startmindmap\n* First\n- Second\n@endmindmap\n"
	want := "mindmap\n  root((First))\n    Second\n"
	got, err := PlantUMLToMermaid(src)
	if err != nil {
		t.Fatalf("PlantUMLToMermaid returned error: %v", err)
	}
	if got != want {
		t.Errorf("PlantUMLToMermaid(%q) = %q, want %q", src, got, want)
	}
}

func TestPlantUMLToMermaidErrors(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"empty", ""},
		{"sequence diagram", "@startuml\nAlice -> Bob: hi\n@enduml\n"},
		{"class diagram", "@startuml\nclass A\n@enduml\n"},
		{"salt diagram", "@startsalt\n{ T\n}\n@endsalt\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PlantUMLToMermaid(tt.src)
			if err == nil {
				t.Fatalf("PlantUMLToMermaid(%q) = %q, want error", tt.src, got)
			}
			if !strings.Contains(err.Error(), "plantuml: unsupported diagram") {
				t.Errorf("PlantUMLToMermaid(%q) error = %q, want %q",
					tt.src, err.Error(), "plantuml: unsupported diagram")
			}
		})
	}
}

// TestPlantUMLToMermaidNeverPanics feeds malformed input to the transpiler; a
// panic fails the test.
func TestPlantUMLToMermaidNeverPanics(t *testing.T) {
	inputs := []string{
		"",
		"@startgantt",
		"@startgantt\n[\n@endgantt",
		"@startgantt\n[] lasts",
		"@startgantt\n[] starts after []",
		"@startgantt\nProject starts\n@endgantt",
		"@startgantt\n-- --",
		"@startmindmap",
		"@startmindmap\n*\n@endmindmap",
		"@startmindmap\n**\n@endmindmap",
		"@startmindmap\n*((\n@endmindmap",
		"\x00@startgantt",
	}
	for _, in := range inputs {
		_, _ = PlantUMLToMermaid(in)
	}
}
