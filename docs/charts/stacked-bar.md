# Stacked bar chart

A vertical bar chart whose series stack; put `"type": "stacked_bar"` in the
spec. Aliases: `stackedbar`, `stacked`.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `stacked_bar` (aliases: `stackedbar`, `stacked`) |
| Golden fixture | none yet |

## Spec

```json
{
  "type": "stacked_bar",
  "labels": ["Mon", "Tue", "Wed"],
  "names": ["CI", "Deploy"],
  "series": [[30, 40, 35], [10, 12, 9]]
}
```

## Output

Stacking is a render option, not a spec key: the adapter feeds `series` through
the engine's `StackSeries`. Every stacked segment is its own filled `<path>`, so
each segment is a mark (`data-role="mark"`, class `svg-mark`).

## Theming

Each segment's fill becomes `var(--diagram-series-N, <engine colour>)`, declared
as `--diagram-series-0`…`-7` on `.svg-diagram` by `DefaultCSS()`; `--diagram-fg`,
`--diagram-muted`, `--diagram-border` and `--diagram-line` style axes.

## Caveats

`labels` names the marks; without it each falls back to an index.

> Coverage: no golden fixture yet; see `BACKLOG.md`.
