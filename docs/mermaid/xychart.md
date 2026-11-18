# XY chart

An XY chart plots bar or line series against labelled axes, opened with the `xychart-beta` keyword.

| | |
| --- | --- |
| Fence | ` ```mermaid ` (alias ` ```mmd `) |
| Engine | pele-go |
| Keyword | `xychart-beta` |
| Golden fixture | none yet |

## Example

```mermaid
xychart-beta
  x-axis [Jan, Feb, Mar]
  y-axis "Revenue" 0 --> 100
  bar [10, 40, 60]
```

## Output

One `<figure class="svg-diagram" data-lang="mermaid" data-engine="pele" data-theme="auto" data-hash="...">` around one inline `<svg viewBox="..." class="svg-diagram-svg" role="img">`. Elements pele marks with its own id get `data-item` and a `data-role` taken from pele's own class, echoed as a class `svg-<role>`; drawing primitives nested inside a stamped item are not stamped again. Stamped groups also get a `<title>` child for zero-JS hover tooltips.

## Theming

Bars, line paths, axes and labels read the shared palette — `--diagram-surface`, `--diagram-border`, `--diagram-fg`, `--diagram-line`, `--diagram-accent` — because pele's `--pele-*` variables are aliased onto `--diagram-*` on the root `<svg>`. Without `DefaultCSS()` pele's own colours are the `var()` fallback.

## Caveats

An unsupported construct renders the 600x120 `svg-diagram-error` card instead of breaking conversion; only `ErrorFail` makes `Markdown.Convert()` return an error. Unlike the other types, the bars and line of an XY chart carry no id in pele's output, so nothing is stamped with `data-item` or `data-role`: an XY chart is not addressable by a client-side handler yet.

> Coverage: no golden fixture yet; see `BACKLOG.md`.
