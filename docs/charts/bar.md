# Bar chart

A grouped vertical bar chart; put `"type": "bar"` in the spec. No aliases.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `bar` (aliases: none) |
| Golden fixture | `testdata/input/bar.chart`; `echarts` alias: `testdata/input/bar-echarts.echarts` |

## Spec

```json
{
  "type": "bar",
  "title": "Requests per day",
  "labels": ["Mon", "Tue", "Wed", "Thu", "Fri"],
  "names": ["Reqs", "Errors"],
  "series": [[120, 200, 150, 80, 70], [10, 20, 5, 8, 3]]
}
```

## Output

Every bar is a filled `<path>`, so each bar is a mark with `data-role="mark"`,
class `svg-mark`, a colour-keyed `data-series` (`0`, `1`, … in `series` order)
and `data-item` from `labels` (the Nth mark of a series is the Nth category).

## Theming

Each mark's fill becomes `var(--diagram-series-N, <engine colour>)`, declared as
`--diagram-series-0`…`-7` on `.svg-diagram` by `DefaultCSS()`; `--diagram-fg`,
`--diagram-muted`, `--diagram-border` and `--diagram-line` style axes.

## Caveats

Marks are named only from `labels`; otherwise each falls back to a positional
index. `names` feeds the legend only and never names a mark.
