# Horizontal bar chart

The bars rotated so categories run down the left; put `"type": "horizontal_bar"`
in the spec. Aliases: `horizontalbar`, `hbar`.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `horizontal_bar` (aliases: `horizontalbar`, `hbar`) |
| Golden fixture | `testdata/input/horizontal-bar.chart` |

## Spec

```json
{
  "type": "horizontal_bar",
  "labels": ["Chrome", "Safari", "Firefox"],
  "series": [[63, 20, 8]]
}
```

## Output

Every bar is a filled `<path>`, so each bar is a mark (`data-role="mark"`, class
`svg-mark`) with a colour-keyed `data-series` and `data-item` from `labels` (the
Nth mark of a series is the Nth category).

## Theming

Each mark's fill becomes `var(--diagram-series-N, <engine colour>)`, declared as
`--diagram-series-0`…`-7` on `.svg-diagram` by `DefaultCSS()`; `--diagram-fg`,
`--diagram-muted`, `--diagram-border` and `--diagram-line` style axes.

## Caveats

`labels` names the marks; without it each mark is named by its 1-based category
number, the same number the axis prints.
