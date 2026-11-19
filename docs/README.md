# goldmark-diagrams documentation

`goldmark-diagrams` is a 100% pure-Go [Goldmark](https://github.com/yuin/goldmark)
extension that compiles fenced diagram and chart code blocks into one
standardized, responsive, CSS-interactive `<svg>`. Every fence renders into a
single `<figure class="svg-diagram">` wrapping one inline `<svg viewBox="...">`,
with `data-*` hooks and `svg-*` classes so a handler or a stylesheet can address
any diagram the same way.

Start with [getting started](./getting-started.md) if you have not wired the
extension into a Go program yet.

## Language guides

| Fence language | Preprocessor | Engine | Guide |
| --- | --- | --- | --- |
| `mermaid`, `mmd` | none | pele-go | [Mermaid types](#mermaid-types) |
| `d2` | none | terrastruct/d2 (dagre) | [D2](./languages/d2.md) |
| `plantuml`, `puml` | PlantUML transpiler | d2, or pele-go for gantt and mindmap | [PlantUML](./languages/plantuml.md) |
| `dot`, `graphviz` | DOT transpiler | terrastruct/d2 | [Graphviz DOT](./languages/dot.md) |
| `chart`, `echarts` | JSON / YAML spec | go-analyze/charts | [Chart types](#chart-types) |

## Mermaid types

Every Mermaid type below is implemented by pele-go and passes straight through.
Each page shows the fence, an example, the hooks the output carries, and the
known limits.

| Type | Page | Type | Page | Type | Page |
| --- | --- | --- | --- | --- | --- |
| Agent flow | [agentflow](./mermaid/agentflow.md) | Info | [info](./mermaid/info.md) | Sankey | [sankey](./mermaid/sankey.md) |
| Architecture | [architecture](./mermaid/architecture.md) | Ishikawa | [ishikawa](./mermaid/ishikawa.md) | Sequence | [sequence](./mermaid/sequence.md) |
| Block | [block](./mermaid/block.md) | Journey | [journey](./mermaid/journey.md) | State | [state](./mermaid/state.md) |
| C4 | [c4](./mermaid/c4.md) | Kanban | [kanban](./mermaid/kanban.md) | Swimlane | [swimlane](./mermaid/swimlane.md) |
| Class | [class](./mermaid/class.md) | Mindmap | [mindmap](./mermaid/mindmap.md) | Timeline | [timeline](./mermaid/timeline.md) |
| Cynefin | [cynefin](./mermaid/cynefin.md) | Packet | [packet](./mermaid/packet.md) | Treemap | [treemap](./mermaid/treemap.md) |
| ER | [er](./mermaid/er.md) | Pie | [pie](./mermaid/pie.md) | Tree view | [treeview](./mermaid/treeview.md) |
| Event modeling | [eventmodeling](./mermaid/eventmodeling.md) | Quadrant | [quadrant](./mermaid/quadrant.md) | Use case | [usecase](./mermaid/usecase.md) |
| Flowchart | [flowchart](./mermaid/flowchart.md) | Radar | [radar](./mermaid/radar.md) | Venn | [venn](./mermaid/venn.md) |
| Gantt | [gantt](./mermaid/gantt.md) | Railroad | [railroad](./mermaid/railroad.md) | Wardley | [wardley](./mermaid/wardley.md) |
| Git graph | [git](./mermaid/git.md) | Requirement | [requirement](./mermaid/requirement.md) | XY chart | [xychart](./mermaid/xychart.md) |

## Chart types

| Type | Page | Type | Page | Type | Page |
| --- | --- | --- | --- | --- | --- |
| Bar | [bar](./charts/bar.md) | Radar | [radar](./charts/radar.md) | Heatmap | [heatmap](./charts/heatmap.md) |
| Stacked bar | [stacked-bar](./charts/stacked-bar.md) | Funnel | [funnel](./charts/funnel.md) | Table | [table](./charts/table.md) |
| Horizontal bar | [horizontal-bar](./charts/horizontal-bar.md) | Pie | [pie](./charts/pie.md) | Scatter | [scatter](./charts/scatter.md) |
| Line | [line](./charts/line.md) | Donut | [donut](./charts/donut.md) | Candlestick | [candlestick](./charts/candlestick.md) |

## Contract and behavior

| Topic | Page |
| --- | --- |
| Installing, wiring the extension, serving the stylesheet | [getting started](./getting-started.md) |
| `data-*` hooks and `svg-*` classes | [hooks](./hooks.md) |
| Palette, `DefaultCSS()`, dark mode | [theming](./theming.md) |
| Server round-trip for kanban and chart drill-downs | [interactivity](./interactivity.md) |
| Never-fail behavior and the error card | [errors](./errors.md) |
| Content-addressed caching | [caching](./caching.md) |
| Options and `RenderStandalone` | [API](./api.md) |

[`../spec.md`](../spec.md) is the authoritative target.
