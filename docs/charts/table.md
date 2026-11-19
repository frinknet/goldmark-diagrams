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

No cell is a chart mark. The header band and the alternating row backgrounds are
filled `<path>` bands, and each one is hooked as a row: `data-role="row"` with
`data-item` naming the row. Cells are the `<text>` elements.

| Element | Hook |
| --- | --- |
| Header band | `data-item="header"` + `data-role="row"` |
| Body row band | `data-item="<first cell>"` + `data-role="row"` |
| Cell text | `data-item="<column header>"` + `data-row="<row item>"` + `data-role="cell"` |

A row with no first cell is named by its 1-based row number. The `svg-row` and
`svg-cell` classes are emitted alongside the roles, so a handler can attach to a
row or a cell without parsing the SVG.

## Theming

The row-strip fills become `var(--diagram-series-N, <engine colour>)`;
`DefaultCSS()` declares `--diagram-series-0`…`-7` on `.svg-diagram`.

## Caveats

`table` is required (otherwise the error card). `width` sets the painter width and
the height is computed from the rows, so the spec's `height` has no effect;
`title`, `labels` and `names` are ignored.
