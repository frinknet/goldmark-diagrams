# Table chart

A plain data table rendered as SVG; put `"type": "table"` in the spec.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `table` (aliases: none) |
| Golden fixture | `testdata/input/table.chart` |

## Spec

```json
{
  "type": "table",
  "table": {
    "header": ["Name", "Age", "Action"],
    "rows": [["John", "32", "Send Mail"], ["Jim", "42", "Send Mail"]]
  }
}
```

## Output

Table content is text, so no cell is a mark. The only filled `<path>` elements
are the header band and alternating row backgrounds; these are stamped with the
generic mark hook but are row strip backgrounds, not data points.

## Theming

The row-strip fills become `var(--diagram-series-N, <engine colour>)`;
`DefaultCSS()` declares `--diagram-series-0`…`-7` on `.svg-diagram`.

## Caveats

`table` is required (otherwise the error card). `width` sets the painter width and
the height is computed from the rows, so the spec's `height` has no effect;
`title`, `labels` and `names` are ignored.
