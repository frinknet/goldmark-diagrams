# Errors

Markdown conversion never fails on a diagram error. Invalid syntax, transpile
failures, and layout timeouts render into the document; the only mode that
returns a Go error is the opt-in `ErrorFail`.

## Error modes

| Mode | Value | Behavior |
| --- | --- | --- |
| `ErrorSVGCard` (default) | `svg_card` | inline `<svg>` error card |
| `ErrorCodeBlock` | `code_block` | `<pre><code class="language-...">` fallback |
| `ErrorFail` | `fail` | `Convert` returns the wrapped error |

```go
md := goldmark.New(goldmark.WithExtensions(diagrams.New(
    diagrams.WithErrorMode(diagrams.ErrorCodeBlock),
)))
```

## The SVG error card

`ErrorSVGCard` produces a valid inline `<svg viewBox="0 0 600 120"
class="svg-diagram-svg svg-diagram-error" role="img">` wrapped in the normal
`<figure>`. It shows:

- the fence language, plus the source line when the fence offset is recoverable
  (`diagram failed: mermaid, line 3`),
- the error message, split across up to two card lines.

```html
<svg viewBox="0 0 600 120" class="svg-diagram-svg svg-diagram-error" role="img"
     aria-label="mermaid diagram error">
  ...
</svg>
```

## Code-block fallback

`ErrorCodeBlock` returns the fence unchanged as a code block:

```html
<pre><code class="language-mermaid">flowchart TD
  A[[
</code></pre>
```

## Fail mode

`ErrorFail` makes `Convert` (and `RenderStandalone`) return
`diagrams: <lang>: <err>`, wrapped so hosts can identify it.

## When errors occur

- an unsupported or malformed chart spec (`type` missing, unknown type, bad JSON),
- a Mermaid / D2 / DOT / PlantUML parse or layout failure,
- a canceled or expired context (`WithTimeout`, default `5s`).

Line numbers are reported when the fence's byte offset is recoverable, including
CRLF input and a fence at offset 0. D2's compiler has no internal cancellation
checkpoints, so a pathological D2 source can outlive the deadline; the renderer
stops waiting when the context ends and renders an error card, though the
compile may continue in the background.

## See also

- [Caching](./caching.md)
- [API](./api.md)
- [Getting started](./getting-started.md)
- [Interactivity](./interactivity.md)
