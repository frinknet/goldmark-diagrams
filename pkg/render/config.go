// Package render is the diagram render pipeline: it resolves a configuration,
// dispatches a fence to the right engine, normalizes the engine SVG into the
// single output contract, and wraps it in the <figure> fragment.
package render

import (
	"time"

	"github.com/frinknet/goldmark-diagrams/pkg/cache"
	"github.com/frinknet/goldmark-diagrams/pkg/engines"
)

// ErrorMode selects what happens when a diagram fails to compile.
type ErrorMode string

const (
	// ErrorSVGCard (default) renders a valid inline <svg> error card.
	ErrorSVGCard ErrorMode = "svg_card"
	// ErrorCodeBlock falls back to a standard <pre><code class="language-..."> block.
	ErrorCodeBlock ErrorMode = "code_block"
	// ErrorFail returns the Go error from markdown.Convert().
	ErrorFail ErrorMode = "fail"
)

// Config holds every tunable of the extension. The public diagrams package
// aliases it, so hosts can inspect a resolved configuration, but should build
// it via Options.
type Config struct {
	Theme     string
	Cache     cache.Cache
	ErrorMode ErrorMode
	Languages []string
	Timeout   time.Duration
	D2Layout  string // "dagre" (default). "elk" requires an external binary and is rejected.
	ID        string // data-id stamped on <figure> for standalone renders.
}

// Option configures New() and RenderStandalone().
type Option func(*Config)

// Default resolves the configuration for a set of options.
func Default(opts ...Option) *Config {
	c := &Config{
		Theme:     "auto",
		ErrorMode: ErrorSVGCard,
		Timeout:   5 * time.Second,
		D2Layout:  "dagre",
	}
	for _, o := range opts {
		o(c)
	}
	if c.Cache == nil {
		c.Cache = cache.NewMemoryLRU(512)
	}
	if len(c.Languages) == 0 {
		c.Languages = engines.SupportedLanguages()
	}
	if c.D2Layout != "dagre" {
		// elk/taro require external plugin binaries, forbidden by design.
		c.D2Layout = "dagre"
	}
	return c
}

// Handles reports whether a fence language is intercepted by this config.
func (c *Config) Handles(lang string) bool {
	canonical := engines.NormalizeLang(lang)
	if canonical == "" {
		return false
	}
	for _, l := range c.Languages {
		if engines.NormalizeLang(l) == canonical {
			return true
		}
	}
	return false
}
