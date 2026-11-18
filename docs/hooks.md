# Hooks

Every primary element of every diagram type is stamped with the same hook
vocabulary, so one client-side handler and one stylesheet can address a Mermaid
node, a D2 shape, a chart mark, or a Kanban card the same way.

## data-* attributes

| Attribute | Meaning |
| --- | --- |
| `data-item` | identifies a primary element (node, shape, mark, task, ...) |
| `data-role` | the element's role name (see the vocabulary below) |
| `data-series` | `0..7` series slot on multi-series chart marks |
| `data-column-id` | Kanban column slug |
| `data-card-id` | Kanban card identity |

Fence attributes are also copied onto the `<figure>` as `data-<key>`
(`id=brd_123` → `data-id="brd_123"`, `resource=orders` →
`data-resource="orders"`).

## Roles

A role is pele's own class remainder for Mermaid, or `shape` / `connection` /
`mark` / `column` / `card` for the other engines, so the vocabulary follows the
engine rather than a fixed list. The roles seen across the common types are
`node`, `edge`, `cluster`, `commit`, `branch`, `task`, `slice`, `point`,
`venn-set`, `venn-region`, `shape`, `connection`, `mark`, `column`, `card`.

Types that name their own shapes add their own roles: `lifeline` (sequence),
`lane` (swimlane), `rule` (railroad), `curve` (radar), `quadrant`, `legend-item`,
`sankey-node`, `sankey-link`, `cynefin-domain`, `cynefin-region`,
`ishikawa-head`, `ishikawa-category`.

Each role is also emitted as a class `svg-<role>`: `svg-node`, `svg-edge`,
`svg-lifeline`, `svg-sankey-node`, and so on. Chart marks also carry
`svg-series-0` … `svg-series-7`.

## Per-engine hooks

| Fence | Hooks |
| --- | --- |
| `mermaid`, `mmd` | `data-item` + `data-role` using pele's own ids; classes `svg-<role>` |
| `d2`, `dot`, `plantuml` | `data-item` + `data-role="shape"` / `"connection"`; classes `svg-shape` / `svg-connection` |
| `chart`, `echarts` | `data-item` (the category label) + `data-series="0..7"` + `data-role="mark"`; classes `svg-mark`, `svg-series-0..7` |
| kanban | `data-column-id` + `data-role="column"`; `data-card-id` + `data-role="card"` |

A hook is a real identity, not a counter: Mermaid elements use pele's ids,
chart marks use the category label, D2 shapes use their label, and Kanban uses
the column/card name. Indices are only a fallback where the engine exposes
nothing else (D2 connections, chart marks with no labels).

## Titles

A `<title>` child (a zero-JS hover tooltip) is added where the engine provides a
label: pele items and D2 shapes get one. Chart marks do not, because the charts
painter exposes no per-mark label.

## Example

```html
<g class="pele-node svg-node" data-id="A" data-item="A" data-role="node">
  <title>A</title>
  ...
</g>
```

## See also

- [Theming](./theming.md)
- [Interactivity](./interactivity.md)
- [D2](./languages/d2.md)
- [API](./api.md)
