# Git graph

A git graph shows commits and branches in a repository's history, opened with the `gitGraph` keyword.

| | |
| --- | --- |
| Fence | ` ```mermaid ` (alias ` ```mmd `) |
| Engine | pele-go |
| Keyword | `gitGraph` |
| Golden fixture | `testdata/input/gitgraph.mermaid` |

## Example

```mermaid
gitGraph
  commit id: "a1"
  branch develop
  commit id: "b2"
  checkout main
  commit id: "c3"
```

## Output

One `<figure class="svg-diagram" data-lang="mermaid" data-engine="pele" data-theme="auto" data-hash="...">` around one inline `<svg viewBox="..." class="svg-diagram-svg" role="img">`. Commits carry `data-item` + `data-role="commit"` (class `svg-commit`); branches carry `data-item` + `data-role="branch"` (class `svg-branch`). Stamped groups get a `<title>` child for zero-JS hover tooltips.

## Theming

Commit dots, branch rails and labels read the shared palette — `--diagram-surface`, `--diagram-border`, `--diagram-fg`, `--diagram-line`, `--diagram-accent` — because pele's `--pele-*` variables are aliased onto `--diagram-*` on the root `<svg>`. Without `DefaultCSS()` pele's own colours are the `var()` fallback.

## Caveats

An unsupported construct renders the 600x120 `svg-diagram-error` card instead of breaking conversion; only `ErrorFail` makes `Markdown.Convert()` return an error.
