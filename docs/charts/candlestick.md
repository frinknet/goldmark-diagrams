# Candlestick chart

An OHLC candlestick chart; put `"type": "candlestick"` in the spec.

| | |
| --- | --- |
| Fence | ` ```chart ` (alias ` ```echarts `) |
| Engine | go-analyze/charts |
| Spec `type` | `candlestick` (aliases: `candle`, `k`) |
| Golden fixture | `testdata/input/candlestick.chart` |

## Spec

```json
{
  "type": "candlestick",
  "labels": ["Mon", "Tue", "Wed"],
  "candles": [[{"open":20,"high":26,"low":18,"close":24},
               {"open":24,"high":25,"low":19,"close":20},
               {"open":20,"high":27,"low":20,"close":26}]]
}
```

## Output

This type reads `candles`, not `series`: one array per series of
`{open, high, low, close}` objects. Candle bodies are filled `<path>` elements,
so each body is a mark (`data-role="mark"`, class `svg-mark`); the high-low wick
is stroked with `fill:none`, so wicks are not marks. `data-item` cycles through
`labels` per colour slot, so rising and falling bodies can repeat a label
instead of naming every candle uniquely.

## Theming

Each body's fill becomes `var(--diagram-series-N, <engine colour>)`;
`DefaultCSS()` declares `--diagram-series-0`…`-7` on `.svg-diagram`.

## Caveats

`candles` is required; a `series`-only spec renders the error card.
