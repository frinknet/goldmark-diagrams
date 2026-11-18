# Theming

One palette, `--diagram-*`, styles every engine. The default values are declared
on `.svg-diagram` by `DefaultCSS()`; each engine's paint is rewritten to
`var(--diagram-*, <engine colour>)`, so the engine's own colour is the fallback
when the stylesheet is not loaded.

## Palette

| Variable | Default (light) |
| --- | --- |
| `--diagram-bg` | `transparent` |
| `--diagram-fg` | `#1e293b` |
| `--diagram-muted` | `#64748b` |
| `--diagram-border` | `#cbd5e1` |
| `--diagram-line` | `#94a3b8` |
| `--diagram-surface` | `#ffffff` |
| `--diagram-surface-alt` | `#f1f5f9` |
| `--diagram-accent` | `#5470c6` |
| `--diagram-series-0` … `--diagram-series-7` | chart palette |
| `--diagram-error-bg` / `-border` / `-fg` | `#fef2f2` / `#fca5a5` / `#b91c1c` |

## Dark mode

Dark values are applied two ways: automatically under
`@media (prefers-color-scheme: dark)`, and explicitly by `data-theme="dark"` on
the `<figure>` (set with `WithTheme("dark")`).

## Override

```css
.svg-diagram {
  --diagram-accent: #7c3aed;
  --diagram-series-0: #7c3aed;
}
```

## How each engine maps onto the palette

- **Mermaid (pele)**: pele's `--pele-*` variables are aliased onto `--diagram-*`
  on the root `<svg>`, for example `--pele-fg:var(--diagram-fg,#1f1f1f)`.
- **D2**: the embedded palette rules (`fill-N1..N7`, `fill-B*`/`AA*`/`AB*`) are
  rewritten by class role to `var(--diagram-*, <hex>)`.
- **Charts**: each mark's inline fill becomes
  `var(--diagram-series-N, <engine colour>)`.

The theme also selects the engine theme: D2 `DarkMauve`, the pele `dark`
Mermaid theme, and the charts `ThemeDark`.

## WithTheme

```go
md := goldmark.New(goldmark.WithExtensions(diagrams.New(
    diagrams.WithTheme("dark"),
)))
```

Accepted values are `auto`, `light`, `dark`, `neutral`. An unknown value is not
rejected; it falls through to each engine's own default (tracked in
[`../BACKLOG.md`](../BACKLOG.md)).

Info-string keys are copied to the figure as `data-<key>`, so a `theme=dark` key
in a fence both selects the theme used for rendering and becomes the figure's
`data-theme`. The reserved keys `lang`, `engine`, `theme`, and `hash` are never
duplicated: the figure keeps exactly one of each.

## See also

- [Hooks](./hooks.md)
- [Interactivity](./interactivity.md)
- [Errors](./errors.md)
- [API](./api.md)
