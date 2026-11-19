# Mindmap

A mindmap radiates topics from a central root, opened with the `mindmap` keyword.

| | |
| --- | --- |
| Fence | ` ```mermaid ` (alias ` ```mmd `) |
| Engine | pele-go |
| Keyword | `mindmap` |
| Golden fixture | `testdata/input/mindmap.mermaid` |

## Example

```mermaid
mindmap
  root((goldmark-diagrams))
    pure Go
      pele-go
    zero CGo
```

## Output

One `<figure class="svg-diagram" data-lang="mermaid" data-engine="pele" data-theme="auto" data-hash="...">` around one inline `<svg viewBox="..." class="svg-diagram-svg" role="img">`. Nodes carry `data-item` + `data-role="node"` (class `svg-node`); stamped groups get a `<title>` child for zero-JS hover tooltips.

## Theming

Node shapes, their borders and the branch lines read the shared palette — `--diagram-surface`, `--diagram-border`, `--diagram-fg`, `--diagram-line`, `--diagram-accent` — because pele's `--pele-*` variables are aliased onto `--diagram-*` on the root `<svg>`. Without `DefaultCSS()` pele's own colours are the `var()` fallback.

## Caveats

An unsupported construct renders the 600x120 `svg-diagram-error` card instead of breaking conversion; only `ErrorFail` makes `Markdown.Convert()` return an error. Mindmap connectors carry no pele id, so each is named after the branch and child it joins (`goldmark-diagrams->pure Go`, class `svg-edge`).
