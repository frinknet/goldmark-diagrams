# Funnel chart

A descending funnel, one segment per value; put `"type": "funnel"` in the spec.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `funnel` (aliases: none) |
| Golden fixture | `testdata/input/funnel.chart` |

## Spec

```json
{
  "type": "funnel",
  "labels": ["Visit", "Sign up", "Purchase"],
  "series": [[1000, 250, 60]]
}
```

## Output

Only `series[0]` is drawn; later series are ignored. Each funnel segment is a
filled `<path>`, so every segment is a mark (`data-role="mark"`, class
`svg-mark`). Segments are coloured one by one, so each takes its own
`data-series` slot and the label lookup restarts per segment.

## Theming

Each segment's fill becomes `var(--diagram-series-N, <engine colour>)`, declared
as `--diagram-series-0`…`-7` on `.svg-diagram` by `DefaultCSS()`; `--diagram-fg`,
`--diagram-muted`, `--diagram-border` and `--diagram-line` style the labels.

## Caveats

`labels` drive `data-item`; each segment is its own colour slot, so segments are
not individually named.
