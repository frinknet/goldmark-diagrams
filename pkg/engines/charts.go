package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-analyze/charts"
	"gopkg.in/yaml.v3"
)

// chartSpec is the JSON/YAML schema accepted by ```chart and ```echarts fences.
//
//	{
//	  "type":  "line|bar|stacked_bar|horizontal_bar|pie|donut|radar|funnel|scatter|candlestick|heatmap|table",
//	  "title": "optional heading",
//	  "labels": ["Mon","Tue",...],            // x-axis categories / pie slice names
//	  "series": [[1,2,3],[4,5,6]],            // one array per series (pie/funnel: single array)
//	  "names":  ["Reqs","Ops"],               // optional series names (legend)
//	  "candles": [[{open,high,low,close},..]],// required by candlestick
//	  "width": 900, "height": 480             // optional canvas size
//	}
type chartSpec struct {
	Type    string      `json:"type" yaml:"type"`
	Title   string      `json:"title,omitempty" yaml:"title,omitempty"`
	Labels  []string    `json:"labels,omitempty" yaml:"labels,omitempty"`
	Series  [][]float64 `json:"series" yaml:"series"`
	Names   []string    `json:"names,omitempty" yaml:"names,omitempty"`
	Width   int         `json:"width,omitempty" yaml:"width,omitempty"`
	Height  int         `json:"height,omitempty" yaml:"height,omitempty"`
	Candles [][]candle  `json:"candles,omitempty" yaml:"candles,omitempty"`
	Table   *tableSpec  `json:"table,omitempty" yaml:"table,omitempty"`
}

type candle struct {
	Open  float64 `json:"open" yaml:"open"`
	High  float64 `json:"high" yaml:"high"`
	Low   float64 `json:"low" yaml:"low"`
	Close float64 `json:"close" yaml:"close"`
}

type tableSpec struct {
	Header []string   `json:"header" yaml:"header"`
	Rows   [][]string `json:"rows" yaml:"rows"`
}

func parseChartSpec(source string) (*chartSpec, error) {
	s := strings.TrimSpace(source)
	spec := &chartSpec{}
	var err error
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		err = json.Unmarshal([]byte(s), spec)
	} else {
		err = yaml.Unmarshal([]byte(s), spec)
	}
	if err != nil {
		return nil, fmt.Errorf("chart spec: %w", err)
	}
	if spec.Type == "" {
		return nil, fmt.Errorf("chart spec: missing \"type\"")
	}
	return spec, nil
}

// renderCharts compiles a chart spec with go-analyze/charts' pure-Go SVG painter.
// The library's real API is charts.XxxRender(...) plus Painter.Bytes(); the
// spec's charts.SVGOutput() does not exist, SVGOutputOptionFunc does.
func renderCharts(_ context.Context, meta Meta, source string) (*RenderResult, error) {
	spec, err := parseChartSpec(source)
	if err != nil {
		return nil, err
	}

	w := orDefault(spec.Width, 900)
	h := orDefault(spec.Height, 480)

	theme := charts.GetDefaultTheme()
	if meta.Theme == "dark" {
		theme = charts.GetTheme(charts.ThemeDark)
	}

	opts := []charts.OptionFunc{
		charts.DimensionsOptionFunc(w, h),
		charts.SVGOutputOptionFunc(),
		charts.ThemeOptionFunc(theme),
	}
	if len(spec.Labels) > 0 {
		opts = append(opts, charts.XAxisLabelsOptionFunc(spec.Labels))
	}
	if len(spec.Names) > 0 {
		opts = append(opts, charts.LegendLabelsOptionFunc(spec.Names))
	}
	if spec.Title != "" {
		opts = append(opts, charts.TitleTextOptionFunc(spec.Title))
	}

	var p *charts.Painter
	switch strings.ToLower(spec.Type) {
	case "bar":
		p, err = charts.BarRender(flatSeries(spec), opts...)
	case "stacked_bar", "stackedbar", "stacked":
		stacked := true
		p, err = charts.Render(charts.ChartOption{
			SeriesList:  charts.NewSeriesListBar(flatSeries(spec)).ToGenericSeriesList(),
			StackSeries: &stacked,
		}, opts...)
	case "horizontal_bar", "horizontalbar", "hbar":
		p, err = charts.HorizontalBarRender(flatSeries(spec), opts...)
	case "line":
		p, err = charts.LineRender(flatSeries(spec), opts...)
	case "scatter":
		p, err = charts.ScatterRender(flatSeries(spec), opts...)
	case "pie":
		p, err = charts.PieRender(firstSeries(spec), opts...)
	case "donut", "doughnut", "ring":
		p, err = charts.DoughnutRender(firstSeries(spec), opts...)
	case "radar":
		p, err = charts.RadarRender(flatSeries(spec), radarIndicators(spec, theme)...)
	case "funnel":
		p, err = charts.FunnelRender(firstSeries(spec), opts...)
	case "candlestick", "candle", "k":
		p, err = charts.CandleStickRender(ohlc(spec), opts...)
	case "heatmap", "heat":
		if len(spec.Series) == 0 {
			return nil, fmt.Errorf("chart type heatmap requires \"series\"")
		}
		p, err = heatMap(spec, theme, w, h, opts)
	case "table":
		if spec.Table == nil {
			return nil, fmt.Errorf("chart type table requires a \"table\" object")
		}
		// TableRenderValues drops the painter options, so it always answers PNG.
		// TableOptionRenderDirect is the only path that can ask for SVG.
		p, err = charts.TableOptionRenderDirect(charts.TableChartOption{
			OutputFormat: charts.ChartOutputSVG,
			Theme:        theme,
			Width:        w,
			Header:       spec.Table.Header,
			Data:         spec.Table.Rows,
		})
	default:
		return nil, fmt.Errorf("unsupported chart type %q", spec.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("charts render: %w", err)
	}
	out, err := p.Bytes()
	if err != nil {
		return nil, fmt.Errorf("charts svg output: %w", err)
	}
	svg := string(out)
	if i := strings.Index(svg, "<svg"); i >= 0 {
		svg = svg[i:]
	} else {
		return nil, &engineError{"charts produced no <svg> element"}
	}
	res := &RenderResult{
		SVG:       svg,
		Engine:    EngineCharts,
		Labels:    spec.Labels,
		Names:     spec.Names,
		ChartType: strings.ToLower(spec.Type),
		Width:     float64(p.Width()),
		Height:    float64(p.Height()),
	}
	if spec.Table != nil && res.ChartType == "table" {
		res.TableHeader = spec.Table.Header
		res.TableRows = spec.Table.Rows
	}
	return res, nil
}

// heatMap uses the painter-level heat map API, which has no XxxRender shortcut.
func heatMap(spec *chartSpec, theme charts.ColorPalette, w, h int, opts []charts.OptionFunc) (*charts.Painter, error) {
	p := charts.NewPainter(charts.PainterOptions{
		OutputFormat: charts.ChartOutputSVG,
		Width:        w,
		Height:       h,
		Theme:        theme,
	})
	values := spec.Series
	if len(values) == 1 && len(spec.Labels) == len(values[0]) {
		// A single row with labels reads better as one row per label.
		values = make([][]float64, len(spec.Labels))
		for i, v := range spec.Series[0] {
			values[i] = []float64{v}
		}
	}
	opt := charts.NewHeatMapOptionWithData(values)
	opt.Theme = theme
	opt.Padding = charts.NewBox(30, 30, 30, 30)
	if spec.Title != "" {
		opt.Title = charts.TitleOption{Text: spec.Title, Offset: charts.OffsetCenter.WithTop(charts.PositionTop), Theme: theme}
	}
	if len(spec.Labels) > 0 {
		opt.XAxis.Labels = spec.Labels
	}
	if err := p.HeatMapChart(opt); err != nil {
		return nil, err
	}
	return p, nil
}

func ohlc(spec *chartSpec) [][]charts.OHLCData {
	out := make([][]charts.OHLCData, len(spec.Candles))
	for i, series := range spec.Candles {
		row := make([]charts.OHLCData, len(series))
		for j, c := range series {
			row[j] = charts.OHLCData{Open: c.Open, High: c.High, Low: c.Low, Close: c.Close}
		}
		out[i] = row
	}
	return out
}

func flatSeries(spec *chartSpec) [][]float64 {
	if len(spec.Series) == 0 {
		return [][]float64{{}}
	}
	return spec.Series
}

func firstSeries(spec *chartSpec) []float64 {
	if len(spec.Series) == 0 {
		return nil
	}
	return spec.Series[0]
}

func radarIndicators(spec *chartSpec, theme charts.ColorPalette) []charts.OptionFunc {
	opts := []charts.OptionFunc{
		charts.DimensionsOptionFunc(orDefault(spec.Width, 900), orDefault(spec.Height, 480)),
		charts.SVGOutputOptionFunc(),
		charts.ThemeOptionFunc(theme),
	}
	if len(spec.Names) > 0 {
		opts = append(opts, charts.LegendLabelsOptionFunc(spec.Names))
	}
	if len(spec.Labels) > 0 {
		max := make([]float64, len(spec.Labels))
		for _, s := range spec.Series {
			for i, v := range s {
				if i < len(max) && v > max[i] {
					max[i] = v
				}
			}
		}
		opts = append(opts, charts.RadarIndicatorOptionFunc(spec.Labels, max))
	}
	if spec.Title != "" {
		opts = append(opts, charts.TitleTextOptionFunc(spec.Title))
	}
	return opts
}

func orDefault(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
