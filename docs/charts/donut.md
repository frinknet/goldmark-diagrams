# Donut chart

A pie with a hole in the middle; put `"type": "donut"` in the spec.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `donut` (aliases: `doughnut`, `ring`) |
| Golden fixture | none yet |

## Spec

```json
{
  "type": "donut",
  "title": "Storage",
  "labels": ["Docs", "Media", "Cache"],
  "series": [[50, 30, 20]]
}
```

## Output

Only `series[0]` is drawn; later series are ignored. Each ring segment is a
filled `<path>`, so every segment is a mark (`data-role="mark"`, class
`svg-mark`). Segments are coloured one by one, so each takes its own
`data-series` slot and the label lookup restarts per segment.

## Theming

Each segment's fill becomes `var(--diagram-series-N, <engine colour>)`;
`DefaultCSS()` declares `--diagram-series-0`…`-7` on `.svg-diagram`.

## Caveats

`labels` drive `data-item`; each segment is its own colour slot, so segments are
not individually named.

> Coverage: no golden fixture yet; see `BACKLOG.md`.
