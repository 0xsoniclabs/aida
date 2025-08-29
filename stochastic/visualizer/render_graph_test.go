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
	"testing"

	"github.com/goccy/go-graphviz"
	"github.com/stretchr/testify/assert"
)

func TestVisualizer_renderDotGraph(t *testing.T) {
	expectedHtml := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Test Graph</title>

    <script>
        const dot = ` + "`" + `digraph "" {
	graph [bb="0,0,0,0"];
	node [label="\N"];
}
` + "`" + `;
    </script>
</head>

<body>
    <h1>Test Graph</h1>
    <div id="graph"></div>
    <script type="module">
        import { Graphviz } from "https://cdn.jsdelivr.net/npm/@hpcc-js/wasm/dist/index.js";
        if (Graphviz) {
            const graphviz = await Graphviz.load();
            const svg = graphviz.layout(dot, "svg", "dot");
	    document.getElementById("graph").innerHTML = svg;
        } 
    </script>
</body>
</html>
`
	g := graphviz.New()
	graph, _ := g.Graph()
	output, err := renderDotGraph("Test Graph", g, graph)
	assert.Nil(t, err)
	assert.Equal(t, expectedHtml, output)
}
