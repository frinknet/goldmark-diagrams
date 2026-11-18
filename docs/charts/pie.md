# Pie chart

A single-ring pie; put `"type": "pie"` in the spec. No aliases.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `pie` (aliases: none) |
| Golden fixture | `testdata/input/pie.chart` |

## Spec

```json
{
  "type": "pie",
  "title": "Pets",
  "labels": ["Dogs", "Cats", "Fish"],
  "series": [[40, 30, 30]]
}
```

## Output

Only `series[0]` is drawn; later series are ignored. Each slice is a filled
`<path>`, so every slice is a mark (`data-role="mark"`, class `svg-mark`).
Slices are coloured one by one, so each takes its own `data-series` slot and the
label lookup restarts per slice.

## Theming

Each slice's fill becomes `var(--diagram-series-N, <engine colour>)`, declared as
`--diagram-series-0`…`-7` on `.svg-diagram` by `DefaultCSS()`; `--diagram-fg`,
`--diagram-muted`, `--diagram-border` and `--diagram-line` style the title.

## Caveats

With `labels` declared, every slice resolves to `labels[0]` (the golden fixture
shows all three slices as `data-item="Dogs"`). A plain pie prints no slice names.
