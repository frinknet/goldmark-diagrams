# Radar chart

A spider chart with one polygon per series; put `"type": "radar"` in the spec.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `radar` (aliases: none) |
| Golden fixture | none yet |

## Spec

```json
{
  "type": "radar",
  "labels": ["Speed", "Power", "Range", "Cost"],
  "names": ["Model A", "Model B"],
  "series": [[80, 60, 70, 40], [50, 90, 55, 65]]
}
```

## Output

Every series is a filled (translucent) polygon, so each polygon is a mark
(`data-role="mark"`, class `svg-mark`). The fill colour keys the `data-series`
slot and `data-item` comes from `labels`; `labels` are the indicator names.

## Theming

Each polygon's fill becomes `var(--diagram-series-N, <engine colour>)`, declared
as `--diagram-series-0`…`-7` on `.svg-diagram` by `DefaultCSS()`; `--diagram-fg`,
`--diagram-muted`, `--diagram-border` and `--diagram-line` style the spokes.

## Caveats

The radar option list omits the theme, so `theme=dark` does not recolour it, and
each polygon resolves to `labels[0]` as its name.

> Coverage: no golden fixture yet; see `BACKLOG.md`.
