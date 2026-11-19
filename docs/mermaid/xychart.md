# XY chart

An XY chart plots bar or line series against labelled axes, opened with the `xychart-beta` keyword.

| | |
| --- | --- |
| Fence | ` ```mermaid ` (alias ` ```mmd `) |
| Engine | pele-go |
| Keyword | `xychart-beta` |
| Golden fixture | `testdata/input/xychart.mermaid` |

## Example

```mermaid
xychart-beta
  x-axis [Jan, Feb, Mar]
  y-axis "Revenue" 0 --> 100
  bar [10, 40, 60]
```

## Output

One `<figure class="svg-diagram" data-lang="mermaid" data-engine="pele" data-theme="auto" data-hash="...">` around one inline `<svg viewBox="..." class="svg-diagram-svg" role="img">`. Elements pele marks with its own id get `data-item` and a `data-role` taken from pele's own class, echoed as a class `svg-<role>`; drawing primitives nested inside a stamped item are not stamped again. Stamped groups also get a `<title>` child for zero-JS hover tooltips.

XY charts get hooks even though pele gives their marks no id. Each bar is `data-item="<x-axis category>" data-series="<n>" data-role="bar"` (class `svg-bar`), each line point is the same with `data-role="point"` (class `svg-point`), and the line polyline carries the plot title, or `line` when there is none, as `data-item` with `data-role="line"`. Categories come from the `x-axis [...]` band; a numeric axis range names the marks with their 1-based category number instead.

## Theming

Bars, line paths, axes and labels read the shared palette — `--diagram-surface`, `--diagram-border`, `--diagram-fg`, `--diagram-line`, `--diagram-accent` — because pele's `--pele-*` variables are aliased onto `--diagram-*` on the root `<svg>`. Without `DefaultCSS()` pele's own colours are the `var()` fallback.

## Caveats

An unsupported construct renders the 600x120 `svg-diagram-error` card instead of breaking conversion; only `ErrorFail` makes `Markdown.Convert()` return an error.

> Coverage: `testdata/golden/xychart.html`.
