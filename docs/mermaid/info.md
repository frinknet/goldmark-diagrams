# Info

An info diagram prints the rendering engine's own name and version, opened with the `info` keyword.

| | |
| --- | --- |
| Fence | ` ```mermaid ` (alias ` ```mmd `) |
| Engine | pele-go |
| Keyword | `info` |
| Golden fixture | `testdata/input/info.mermaid` |

## Example

```mermaid
info showInfo
```

## Output

One `<figure class="svg-diagram" data-lang="mermaid" data-engine="pele" data-theme="auto" data-hash="...">` around one inline `<svg viewBox="..." class="svg-diagram-svg" role="img">`. The body is a title and a version line; any stamped group gets a `<title>` child for zero-JS hover tooltips.

The version line is the diagram's one addressable element: its own text is the `data-item` (for example `Pele v1.10.0`) and it carries `data-role="version"` plus class `svg-version`.

## Theming

The title and version text read the shared palette — `--diagram-fg` for the title and `--diagram-muted` for the version, with `--diagram-surface` behind them — because pele's `--pele-*` variables are aliased onto `--diagram-*` on the root `<svg>`. Without `DefaultCSS()` pele's own colours are the `var()` fallback.

## Caveats

An unsupported construct renders the 600x120 `svg-diagram-error` card instead of breaking conversion; only `ErrorFail` makes `Markdown.Convert()` return an error.

> Coverage: `testdata/golden/info.html`.
