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

Each plotted point is a `<circle>` element, not a filled `<path>`. The
normalizer only treats a filled `<path>`/`<rect>` as a mark, so a plain scatter
chart emits no `data-item` / `data-series` hooks. `labels` and `names` still
drive the axes and legend.

## Theming

With no rewritten mark fills, `--diagram-series-N` governs nothing here and the
engine's own colours show through; `--diagram-fg`, `--diagram-muted`,
`--diagram-border` and `--diagram-line` style the axes and labels.

## Caveats

`labels` and `names` feed the axes and legend only; circles carry no per-point
mark handle.
