# API

The public API lives in package `diagrams` (module
`github.com/goldmark-diagrams/goldmark-diagrams`).

## Functions and constants

```go
const ModuleVersion = "v0.3.0"

func DefaultCSS() string
func SupportedLanguages() []string
func New(opts ...Option) goldmark.Extender
func RenderStandalone(ctx context.Context, lang, source string, opts ...Option) (string, error)
```

## ErrorMode

```go
type ErrorMode string

const (
	ErrorSVGCard   ErrorMode = "svg_card"   // default
	ErrorCodeBlock ErrorMode = "code_block"
	ErrorFail      ErrorMode = "fail"
)
```

## Cache

```go
type Cache interface {
	Get(key string) (html string, ok bool)
	Set(key string, html string)
}

func NewMemoryLRU(maxEntries int) Cache  // default 512
func NewDiskCache(dir string) Cache      // atomic temp-file + rename writes
```

## Options

| Option | Default |
| --- | --- |
| `WithTheme("auto"｜"light"｜"dark"｜"neutral")` | `auto` |
| `WithCache(cache)` | `NewMemoryLRU(512)` |
| `WithErrorMode(mode)` | `ErrorSVGCard` |
| `WithLanguages(langs...)` | every supported language |
| `WithTimeout(d)` | `5s` |
| `WithD2Layout("dagre")` | `dagre` (`elk` is rejected) |
| `WithID(id)` | unset |

## Config

`Option`s mutate an exported `Config`, resolved by the defaults above:

```go
type Config struct {
	Theme     string        // "auto" | "light" | "dark" | "neutral"
	Cache     Cache
	ErrorMode ErrorMode
	Languages []string
	Timeout   time.Duration
	D2Layout  string        // "dagre"; any other value falls back to "dagre"
	ID        string        // data-id stamped for standalone renders
}
```

## RenderStandalone

```go
html, err := diagrams.RenderStandalone(ctx, "mermaid", src,
	diagrams.WithID("brd_123"),
	diagrams.WithTheme("dark"),
	diagrams.WithErrorMode(diagrams.ErrorSVGCard),
)
```

It returns a complete `<figure>…</figure>` fragment. With the default
`ErrorSVGCard` mode it never returns an error.

## The output contract

```html
<figure class="svg-diagram"
        data-lang="mermaid"
        data-engine="pele"
        data-theme="auto"
        data-hash="9f86d081"
        data-id="brd_123">
  <svg viewBox="0 0 900 480" class="svg-diagram-svg" role="img" aria-label="mermaid diagram">
    ...
  </svg>
</figure>
```

## See also

- [Getting started](./getting-started.md)
- [Hooks](./hooks.md)
- [Theming](./theming.md)
- [Errors](./errors.md)
- [Caching](./caching.md)
- [Interactivity](./interactivity.md)
