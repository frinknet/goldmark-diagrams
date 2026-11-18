# Line chart

A multi-series line chart; put `"type": "line"` in the spec. No aliases.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `line` (aliases: none) |
| Golden fixture | none yet |

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
The normalizer only treats a filled `<path>`/`<rect>` as a mark, so neither
qualifies: a plain line chart emits no `data-item` / `data-series` marks.

## Theming

With no rewritten mark fills, `--diagram-series-N` governs nothing here and the
engine's own stroke colours show through; `--diagram-fg`, `--diagram-muted`,
`--diagram-border` and `--diagram-line` style the axes and labels.

## Caveats

`names` feeds the legend only; stroked lines carry no per-series mark handle.

> Coverage: no golden fixture yet; see `BACKLOG.md`.
