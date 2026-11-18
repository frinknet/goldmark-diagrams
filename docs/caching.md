# Caching

Every compiled `<figure>` fragment is content-addressed, so the same diagram and
configuration is rendered once and reused. The cache is an optimization: a miss
or a write failure never changes the output.

## The Cache interface

```go
type Cache interface {
	Get(key string) (html string, ok bool)
	Set(key string, html string)
}
```

## Key formula

```
sha256(ModuleVersion ":" lang ":" theme ":" attrs ":" source)
```

The digest is taken over the normalized fragment inputs: the module version, the
canonical fence language, the resolved theme, the fence attributes in stable
order, and the source. Bumping `ModuleVersion` invalidates every cached entry.

## Built-in stores

| Constructor | Behavior |
| --- | --- |
| `NewMemoryLRU(maxEntries int)` | thread-safe in-memory LRU; a non-positive `maxEntries` falls back to `512` |
| `NewDiskCache(dir string)` | one `<hash>.svg` per entry, written atomically via temp file + rename |

`New()` installs `NewMemoryLRU(512)` unless `WithCache` overrides it.

## Wiring

```go
md := goldmark.New(goldmark.WithExtensions(diagrams.New(
    diagrams.WithCache(diagrams.NewDiskCache("/var/cache/diagrams")),
)))
```

The disk cache survives restarts and never exposes a partial file: writes go
through `CreateTemp` followed by `Rename`, and any failure is swallowed.

## What is cached

Both successful fragments and `ErrorSVGCard` output are cached.
`ErrorCodeBlock` and `ErrorFail` are not written to the cache.

## See also

- [API](./api.md)
- [Errors](./errors.md)
- [Getting started](./getting-started.md)
