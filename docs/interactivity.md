# Interactivity

The package ships no JavaScript. Instead, every diagram emits predictable
`data-*` attributes, and pure CSS drives hover dimming, animated dashed
connectors, and dark mode (see [Theming](./theming.md)). A host application can
add a thin delegated listener for server-backed mutations such as Kanban card
moves and chart drill-downs.

## The feedback loop

1. The backend renders the diagram, either as Markdown or with
   `diagrams.RenderStandalone(ctx, "mermaid", source, diagrams.WithID("brd_123"))`.
2. A delegated client listener reads `data-id`, `data-card-id`,
   `data-column-id`, or `data-item` from the clicked element and POSTs JSON to a
   backend endpoint.
3. The backend updates its store and re-renders with `RenderStandalone`.
4. The client swaps `figure.outerHTML` with the fresh fragment, so all state
   stays in the database.

## Illustrative listener (host-provided)

```js
document.addEventListener("click", async (ev) => {
  const el = ev.target.closest("[data-item],[data-card-id]");
  if (!el) return;
  const figure = el.closest("figure.svg-diagram");
  const res = await fetch("/_diagrams/move", {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({
      id: figure.dataset.id,
      card: el.dataset.cardId,
      column: el.dataset.columnId,
    }),
  });
  figure.outerHTML = await res.text();
});
```

## Backend endpoint

```go
html, err := diagrams.RenderStandalone(ctx, "mermaid", kanbanSource,
    diagrams.WithID(brdID), diagrams.WithTheme("auto"))
if err != nil {
    // Only reachable with ErrorFail; the default ErrorSVGCard never returns an error.
}
w.Write([]byte(html))
```

The returned string is a complete `<figure>…</figure>` fragment, ready to
replace `figure.outerHTML`.

## Attributes the listener relies on

| Attribute | Element | Use |
| --- | --- | --- |
| `data-id` | `<figure>` | identifies the whole diagram (fence `id=...` or `WithID`) |
| `data-item` | primary mark | node / shape / chart mark identity |
| `data-row` | table cell | the row a cell belongs to (e.g. for a drill-down) |
| `data-card-id` | kanban card | card identity for a move |
| `data-column-id` | kanban column | destination column slug |

## See also

- [Hooks](./hooks.md)
- [API](./api.md)
- [Errors](./errors.md)
- [Getting started](./getting-started.md)
