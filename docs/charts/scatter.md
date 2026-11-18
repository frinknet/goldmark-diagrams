# Scatter chart

A point cloud per series; put `"type": "scatter"` in the spec. No aliases.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `scatter` (aliases: none) |
| Golden fixture | none yet |

## Spec

```json
{
  "type": "scatter",
  "labels": ["A", "B", "C", "D"],
  "names": ["Run 1", "Run 2"],
  "series": [[5, 9, 3, 7], [2, 4, 8, 1]]
}
```

## Output

Each plotted point is a filled `<path>`, so every point is a mark
(`data-role="mark"`, class `svg-mark`). All points of one series share a colour
and so one `data-series` slot; `data-item` comes from `labels` by position.

## Theming

Each point's fill becomes `var(--diagram-series-N, <engine colour>)`, declared as
`--diagram-series-0`…`-7` on `.svg-diagram` by `DefaultCSS()`; `--diagram-fg`,
`--diagram-muted`, `--diagram-border` and `--diagram-line` style axes.

## Caveats

`labels` names the marks; a dense series yields many `data-item` hooks.

> Coverage: no golden fixture yet; see `BACKLOG.md`.
