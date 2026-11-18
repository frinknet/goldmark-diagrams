# Info

An info diagram prints the rendering engine's own name and version, opened with the `info` keyword.

| | |
| --- | --- |
| Fence | ` ```mermaid ` (alias ` ```mmd `) |
| Engine | pele-go |
| Keyword | `info` |
| Golden fixture | none yet |

## Example

```mermaid
info showInfo
```

## Output

One `<figure class="svg-diagram" data-lang="mermaid" data-engine="pele" data-theme="auto" data-hash="...">` around one inline `<svg viewBox="..." class="svg-diagram-svg" role="img">`. The body is a title and a version line; any stamped group gets a `<title>` child for zero-JS hover tooltips.

## Theming

The title and version text read the shared palette — `--diagram-fg` for the title and `--diagram-muted` for the version, with `--diagram-surface` behind them — because pele's `--pele-*` variables are aliased onto `--diagram-*` on the root `<svg>`. Without `DefaultCSS()` pele's own colours are the `var()` fallback.

## Caveats

An unsupported construct renders the 600x120 `svg-diagram-error` card instead of breaking conversion; only `ErrorFail` makes `Markdown.Convert()` return an error.

> Coverage: no golden fixture yet; see `BACKLOG.md`.
