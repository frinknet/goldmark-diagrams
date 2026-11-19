# Scatter chart

A point cloud per series; put `"type": "scatter"` in the spec. No aliases.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `scatter` (aliases: none) |
| Golden fixture | `testdata/input/scatter.chart` |

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

Each plotted point is a `<circle>` element. The normalizer stamps every point
with `data-item` (the category label, or the 1-based category number when
`labels` is absent), `data-series="0..7"` and `data-role="mark"`, so one
handler can address each point.

## Theming

Each point's fill is rewritten to `var(--diagram-series-N, <engine colour>)`, so
`--diagram-series-N` governs it; `--diagram-fg`, `--diagram-muted`,
`--diagram-border` and `--diagram-line` style the axes and labels.

## Caveats

`labels` and `names` feed the axes and legend; the legend swatches are not
stamped as points.
