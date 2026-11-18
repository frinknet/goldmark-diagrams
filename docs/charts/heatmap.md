# Heatmap chart

A value grid shaded by magnitude; put `"type": "heatmap"` in the spec.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `heatmap` (aliases: `heat`) |
| Golden fixture | `testdata/input/heatmap.chart` |

## Spec

```json
{
  "type": "heatmap",
  "title": "Traffic by hour",
  "labels": ["A", "B", "C"],
  "series": [[1, 5, 3], [4, 2, 9], [7, 1, 6]]
}
```

## Output

`series` is required; without it the spec is rejected. A rows × columns array is
drawn as a grid of cells, each a filled `<path>` whose colour encodes the value,
so every cell is a mark (`data-role="mark"`, class `svg-mark`).

## Theming

Each cell's fill becomes `var(--diagram-series-N, <engine colour>)`, declared as
`--diagram-series-0`…`-7` on `.svg-diagram` by `DefaultCSS()`; `--diagram-fg`,
`--diagram-muted`, `--diagram-border` and `--diagram-line` style the axes.

## Caveats

`names` is ignored; each cell colour is its own slot, so `data-item` cycles
through `labels` by slot rather than identifying a cell's row and column. A
single row whose length equals `len(labels)` is transposed to one row per label.
