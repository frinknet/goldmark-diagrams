# Gantt

A Gantt chart lays out tasks and milestones on a calendar, opened with the `gantt` keyword.

| | |
| --- | --- |
| Fence | ` ```mermaid ` (alias ` ```mmd `) |
| Engine | pele-go |
| Keyword | `gantt` |
| Golden fixture | `testdata/input/gantt.mermaid` |

## Example

```mermaid
gantt
  title Release plan
  dateFormat YYYY-MM-DD
  section Core
  Pipeline :c1, 2024-01-01, 20d
  Engines :c2, after c1, 15d
```

## Output

One `<figure class="svg-diagram" data-lang="mermaid" data-engine="pele" data-theme="auto" data-hash="...">` around one inline `<svg viewBox="..." class="svg-diagram-svg" role="img">`. Tasks carry `data-item` + `data-role="task"` (class `svg-task`); stamped groups get a `<title>` child for zero-JS hover tooltips.

## Theming

Task bars, sections and the grid read the shared palette — `--diagram-surface`, `--diagram-border`, `--diagram-fg`, `--diagram-line`, `--diagram-accent` — because pele's `--pele-*` variables are aliased onto `--diagram-*` on the root `<svg>`. Without `DefaultCSS()` pele's own colours are the `var()` fallback.

## Caveats

An unsupported construct renders the 600x120 `svg-diagram-error` card instead of breaking conversion; only `ErrorFail` makes `Markdown.Convert()` return an error. A Gantt needs sane dates; pele reads them through its own options, so a task with an unparseable start or duration fails to lay out.
