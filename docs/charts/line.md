# Line chart

A multi-series line chart; put `"type": "line"` in the spec. No aliases.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `line` (aliases: none) |
| Golden fixture | `testdata/input/line.chart` |

## Spec

```json
{
  "type": "line",
  "title": "Latency",
  "labels": ["00:00", "06:00", "12:00", "18:00"],
  "names": ["p50", "p99"],
  "series": [[12, 14, 20, 18], [40, 55, 90, 70]]
}
```

## Output

A line is a stroked `<path>` with `fill:none` and its symbols are `<circle>`.
The normalizer stamps the stroked path with `data-item` (the series name, or
`line` when `names` is absent), `data-series="0..7"` and `data-role="line"`,
so one handler can address each series.

## Theming

The line's stroke is rewritten to `var(--diagram-series-N, <engine colour>)`, so
`--diagram-series-N` governs it; `--diagram-fg`, `--diagram-muted`,
`--diagram-border` and `--diagram-line` style the axes and labels.

## Caveats

`names` feeds the legend and the per-series `data-item`; the line's symbols
(`<circle>`) carry no per-point handle.
