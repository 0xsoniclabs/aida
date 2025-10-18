// Copyright 2025 Sonic Labs
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package visualizer

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// HTML references for the rendered pages.
const countingRef = "counting-stats"
const queuingRef = "queuing-stats"
const snapshotRef = "snapshot-stats"
const scalarRef = "scalar-stats"
const operationRef = "operation-stats"
const txoperationRef = "tx-operation-stats"
const simplifiedMarkovRef = "simplified-markov-stats"
const markovRef = "markov-stats"

// MainHtml is the index page.
const MainHtml = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Aida: Stochastic Estimator</title>
    <link rel="stylesheet" href="style.css">
    <script src="script.js"></script>
  </head>
  <body>
    <h1>Aida: Stochastic Estimator</h1>
    <ul>
    <li> <h3> <a href="/` + countingRef + `"> Counting Statistics </a> </h3> </li>
    <li> <h3> <a href="/` + queuingRef + `"> Queuing Statistics </a> </h3> </li>
    <li> <h3> <a href="/` + snapshotRef + `"> Snapshot Statistics </a> </h3> </li>
    <li> <h3> <a href="/` + scalarRef + `"> Scalar Argument Statistics </a> </h3> </li>
    <li> <h3> <a href="/` + txoperationRef + `"> Transactional Operation Statistics  </a> </h3> </li>
    <li> <h3> <a href="/` + operationRef + `"> Operation Statistics  </a> </h3> </li>
    <li> <h3> <a href="/` + simplifiedMarkovRef + `"> Simplified Markov Chain </a> </h3> </li>
    <li> <h3> <a href="/` + markovRef + `"> Markov Chain </a> </h3> </li>
    </ul>
</body>
</html>
`

// renderMain renders the main menu.
func renderMain(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprint(w, MainHtml)
}

// convertCountingData converts CDF points to chart points.
func convertCountingData(data [][2]float64) []opts.LineData {
	items := []opts.LineData{}
	for _, pair := range data {
		items = append(items, opts.LineData{Value: pair})
	}
	return items
}

// newCountingChart creates a line chart for a counting statistic.
func newCountingChart(title string, contracts [][2]float64, keys [][2]float64, values [][2]float64) *charts.Line {
	chart := charts.NewLine()
	chart.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{
		Theme: types.ThemeChalk,
	}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Title: "Save",
				},
				DataZoom: &opts.ToolBoxFeatureDataZoom{
					Show: true,
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{Show: true}),
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}))
	chart.AddSeries("Contracts", convertCountingData(contracts)).AddSeries("Keys", convertCountingData(keys)).AddSeries("Values", convertCountingData(values))

	return chart
}

// renderCounting renders counting statistics.
func renderCounting(w http.ResponseWriter, r *http.Request) {
	view, err := currentView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	stats := view.stats
	chart := newCountingChart(
		"Counting Statistics",
		stats.Contracts.Counting.ECDF,
		stats.Keys.Counting.ECDF,
		stats.Values.Counting.ECDF,
	)
	_ = chart.Render(w)
}

// newScalarChart creates a line chart for scalar argument distributions.
func newScalarChart(balance, nonce, code [][2]float64) *charts.Line {
	chart := charts.NewLine()
	chart.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{
		Theme: types.ThemeChalk,
	}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Title: "Save",
				},
				DataZoom: &opts.ToolBoxFeatureDataZoom{
					Show: true,
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{Show: true}),
		charts.WithTitleOpts(opts.Title{
			Title: "Scalar Argument Statistics",
		}))
	chart.AddSeries("Balance", convertCountingData(balance)).
		AddSeries("Nonce", convertCountingData(nonce)).
		AddSeries("Code Size", convertCountingData(code))
	return chart
}

// renderScalarStats renders scalar argument distributions.
func renderScalarStats(w http.ResponseWriter, r *http.Request) {
	view, err := currentView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	chart := newScalarChart(
		view.stats.Balance.ECDF,
		view.stats.Nonce.ECDF,
		view.stats.CodeSize.ECDF,
	)
	_ = chart.Render(w)
}

// renderSnapshotStats renders a line chart for a snapshot statistics
func renderSnapshotStats(w http.ResponseWriter, r *http.Request) {
	chart := charts.NewLine()
	chart.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{
		Theme: types.ThemeChalk,
	}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Title: "Save",
				},
				DataZoom: &opts.ToolBoxFeatureDataZoom{
					Show: true,
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{Show: true}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Snapshot Statistics",
			Subtitle: "Delta Distribution",
		}))
	view, err := currentView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	chart.AddSeries("eCDF", convertCountingData(view.stats.SnapshotECDF))
	_ = chart.Render(w)
}

// convertQueuingData rendering plot data for the queuing statistics.
func convertQueuingData(data []float64) []opts.ScatterData {
	items := []opts.ScatterData{}
	for x, p := range data {
		items = append(items, opts.ScatterData{Value: [2]float64{float64(x), p}, SymbolSize: 5})
	}
	return items
}

// renderQueuing renders a queuing statistics.
func renderQueuing(w http.ResponseWriter, r *http.Request) {
	view, err := currentView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	stats := view.stats
	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{
		Theme:     types.ThemeChalk,
		PageTitle: "Queuing Probabilities",
	}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Title: "Save",
				},
				DataZoom: &opts.ToolBoxFeatureDataZoom{
					Show: true,
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{Show: true}),
		charts.WithTitleOpts(opts.Title{
			Title: "Queuing Probabilities",
		}))
	scatter.AddSeries("Contract", convertQueuingData(stats.Contracts.Queuing.Distribution)).
		AddSeries("Keys", convertQueuingData(stats.Keys.Queuing.Distribution)).
		AddSeries("Values", convertQueuingData(stats.Values.Queuing.Distribution))
	_ = scatter.Render(w)
}

// convertOperationData produces the data series for the sationary distribution.
func convertOperationData(data []opDatum) []opts.BarData {
	items := []opts.BarData{}
	for i := 0; i < len(data); i++ {
		items = append(items, opts.BarData{Value: data[i].value})
	}
	return items
}

// convertOperationLabel produces operations' labels.
func convertOperationLabel(data []opDatum) []string {
	items := []string{}
	for i := 0; i < len(data); i++ {
		items = append(items, data[i].label)
	}
	return items
}

// renderOperationStats renders the stationary distribution.
func renderOperationStats(w http.ResponseWriter, r *http.Request) {
	view, err := currentView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{
		Theme:     types.ThemeChalk,
		PageTitle: "StateDB Operations",
		Height:    "1300px",
	}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Title: "Save",
				},
				DataZoom: &opts.ToolBoxFeatureDataZoom{
					Show: true,
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{Show: true}),
		charts.WithTitleOpts(opts.Title{
			Title: "StateDB Operations",
		}))
	bar.SetXAxis(convertOperationLabel(view.stationary)).AddSeries("Stationary Distribution", convertOperationData(view.stationary))
	bar.XYReversal()
	_ = bar.Render(w)
}

// renderTransactionalOperationStats renders the average number of operations per transaction.
func renderTransactionalOperationStats(w http.ResponseWriter, r *http.Request) {
	view, err := currentView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	title := fmt.Sprintf("Average %.1f Tx/Bl; %.1f Bl/Ep", view.txPerBlock, view.blocksPerSyncPeriod)
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{
		Theme:     types.ThemeChalk,
		PageTitle: title,
		Height:    "1300px",
	}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Title: "Save",
				},
				DataZoom: &opts.ToolBoxFeatureDataZoom{
					Show: true,
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{Show: true}),
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}))
	bar.SetXAxis(convertOperationLabel(view.txOperation)).AddSeries("Ops/Tx", convertOperationData(view.txOperation))
	_ = bar.Render(w)
}

// printMarkovInDotty renders a markov chain in dotty format
func printMarkovInDotty(title string, stochasticMatrix [][]float64, label []string) (out string, err error) {
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		return "", fmt.Errorf("renderMarkovChain: failed to create graph. Error: %v", err)
	}
	defer func() {
		err = errors.Join(err, graph.Close(), g.Close())
	}()
	n := len(label)
	nodes := make([]*cgraph.Node, n)
	for i := 0; i < n; i++ {
		var err error
		nodes[i], err = graph.CreateNode(label[i])
		if err != nil {
			return "", fmt.Errorf("renderMarkovChain: failed to create node for label (%v, %v). Error: %v", i, label[i], err)
		}
		nodes[i].SetLabel(label[i])
	}
	if n != len(stochasticMatrix) {
		return "", fmt.Errorf("renderMarkovChain: stochastic matrix has %d rows, expected %d", len(stochasticMatrix), n)
	}
	for i := 0; i < n; i++ {
		if stochasticMatrix[i] == nil {
			return "", fmt.Errorf("renderMarkovChain: stochastic matrix row %d is nil", i)
		} else if len(stochasticMatrix[i]) != n {
			return "", fmt.Errorf("renderMarkovChain: stochastic matrix row %d has length %d, expected %d", i, len(stochasticMatrix[i]), n)
		}
		for j := 0; j < n; j++ {
			p := stochasticMatrix[i][j]
			if p > 0.0 {
				txt := fmt.Sprintf("%.2f", p)
				e, _ := graph.CreateEdge("", nodes[i], nodes[j])
				e.SetLabel(txt)
				var color string
				switch int(4 * p) {
				case 0:
					color = "gray"
				case 1:
					color = "green"
				case 3:
					color = "indianred"
				case 4:
					color = "red"
				}
				e.SetColor(color)
			}
		}
	}
	txt, err := renderDotGraph(title, g, graph)
	if err != nil {
		return "", fmt.Errorf("renderMarkovChain: failed to render. Error: %v", err)
	}
	return txt, nil
}

// renderMarkovChain renders a markov chain.
func renderMarkovChain(w http.ResponseWriter, r *http.Request) {
	view, err := currentView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	txt, err := printMarkovInDotty("StateDB Markov-Chain", view.stats.StochasticMatrix, view.stats.Operations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	_, _ = fmt.Fprint(w, txt)
}

// renderSimplifiedMarkovChain renders the simplified markov chain whose nodes have no argument classes.
func renderSimplifiedMarkovChain(w http.ResponseWriter, r *http.Request) {
	view, err := currentView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	label := make([]string, operations.NumOps)
	for i := 0; i < operations.NumOps; i++ {
		label[i] = operations.OpMnemo(i)
	}
	txt, err := printMarkovInDotty("StateDB Simplified Markov-Chain", view.simplifiedMatrix, label)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	_, _ = fmt.Fprint(w, txt)
}

// FireUpWeb produces a data model for the recorded markov stats and
// visualizes with a local web-server.
func FireUpWeb(statsJSON *recorder.StatsJSON, addr string) error {
	if err := setViewState(statsJSON); err != nil {
		return err
	}

	// create web server
	http.HandleFunc("/", renderMain)
	http.HandleFunc("/"+countingRef, renderCounting)
	http.HandleFunc("/"+queuingRef, renderQueuing)
	http.HandleFunc("/"+snapshotRef, renderSnapshotStats)
	http.HandleFunc("/"+scalarRef, renderScalarStats)
	http.HandleFunc("/"+operationRef, renderOperationStats)
	http.HandleFunc("/"+txoperationRef, renderTransactionalOperationStats)
	http.HandleFunc("/"+simplifiedMarkovRef, renderSimplifiedMarkovChain)
	http.HandleFunc("/"+markovRef, renderMarkovChain)
	return http.ListenAndServe(":"+addr, nil)
}
