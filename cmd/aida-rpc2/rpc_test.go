package aidarpc2

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/rpc"
	"github.com/0xsoniclabs/sonic/gossip/filters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebClient_CanConnectAndSendRequest(t *testing.T) {

	url := "http://127.0.0.1:18545"

	client := NewWebClient(url)

	/*
		request := `
		{"method":"eth_getLogs","params":[{"address": "0xeF4B763385838FfFc708000f884026B8c0434275"}],"id":1,"jsonrpc":"2.0"}
		`

		request := `
		{"method":"eth_getLogs","params":[{"fromBlock": "0x3bc2bc4", "toBlock": "0x3bc2bc4"}],"id":1,"jsonrpc":"2.0"}
		`
		request := `
		{"method":"eth_getLogs","params":[{"fromBlock": "0x90", "toBlock": "0x99"}],"id":1,"jsonrpc":"2.0"}
		`
	*/

	request := `
	{"method":"eth_getLogs","params":[{"address":null,"fromBlock":"0x3bc2bb2","toBlock":"0x3bc2bb8","topics":[["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]]}],"id":1,"jsonrpc":"2.0"}
	`

	response, err := client.SendRequest([]byte(request))
	require.NoError(t, err)
	require.NotNil(t, response)

	fmt.Printf("Response: %s\n", string(response))

	require.NoError(t, client.Close())
	t.Fail()

}

func TestWebClient_RunRpcExampleQueries(t *testing.T) {
	rpcLogPath := "/media/herbert/WorkData/sonic/rpc_logs/rpc_logs.gz"
	url := "http://127.0.0.1:18545"

	const numQueries = 10_000

	iter, err := rpc.NewFileReader(context.Background(), rpcLogPath)
	assert.NoError(t, err)
	defer iter.Close()

	client := NewWebClient(url)

	stats := &Stats{}
	timeouts := 0

	c := 0
	for iter.Next() {

		req := iter.Value()
		assert.NoError(t, iter.Error())
		assert.NotNil(t, req)

		if req.Query.Method != "eth_getLogs" {
			continue
		}

		// Skip open-end queries since they may return different results over time.
		if !strings.Contains(string(req.ParamsRaw), "toBlock") {
			continue
		}
		if strings.Contains(string(req.ParamsRaw), `"toBlock":"latest"`) {
			continue
		}

		// Ignore queries with empty results.
		if len(req.ResponseRaw) <= 2 {
			continue
		}

		// Parse query parameters.
		var filterQuery []filters.FilterCriteria
		err = json.Unmarshal(req.ParamsRaw, &filterQuery)
		assert.NoError(t, err)
		require.Equal(t, 1, len(filterQuery))

		// Skip all queries that do not filter by topics or addresses.
		query := filterQuery[0]
		hasFilterCriteria := len(query.Addresses) > 0
		if !hasFilterCriteria {
			for _, topics := range query.Topics {
				if len(topics) > 0 {
					hasFilterCriteria = true
					break
				}
			}
		}

		if !hasFilterCriteria {
			continue
		}

		/*
			fmt.Printf("Request:\n")
			fmt.Printf("  Time: 	 %v\n", time.Unix(int64(req.Timestamp), 0))
			fmt.Printf("  Method:	 %s\n", req.Query.Method)
			fmt.Printf("  Params:	 %s\n", req.Query.Params)
			fmt.Printf("  ParamsRaw: %s\n", string(req.ParamsRaw))
			//fmt.Printf("  Response:	 %s\n", string(req.Response.Result))
		*/

		fmt.Printf("  ParamsRaw: %s\n", string(req.ParamsRaw))
		request := fmt.Sprintf(`{"method":"%s","params":%s,"id":1,"jsonrpc":"2.0"}`, req.Query.Method, string(req.ParamsRaw))

		start := time.Now()
		response, err := client.SendRequest([]byte(request))
		assert.NoError(t, err)
		assert.NotNil(t, response)
		elapsed := time.Since(start)
		stats.Add(elapsed)

		//fmt.Printf("  Response: %s\n", string(response))
		if strings.Contains(string(response), `"error":{"code":-32000,"message":"context deadline exceeded"}`) {
			timeouts++
			continue
			//t.Fatalf("RPC timeout: %s", string(response))
		}

		result := struct {
			Result json.RawMessage `json:"result"`
		}{}
		err = json.Unmarshal(response, &result)
		assert.NoError(t, err)
		assert.NotNil(t, result.Result)

		var want []log
		var got []log

		err = json.Unmarshal(req.ResponseRaw, &want)
		assert.NoError(t, err)

		err = json.Unmarshal(result.Result, &got)
		assert.NoError(t, err)

		//require.Equal(t, len(want), len(got))
		//require.ElementsMatch(t, want, got)
		/*
			for i := range want {
				require.Equal(t, want[i], got[i], "mismatch at index %d", i)
			}
		*/
		//require.Equal(t, want, got)

		//fmt.Printf("  Got: %s\n", string(result.Result))
		//require.Equal(t, req.ResponseRaw, []byte(result.Result))

		c++
		if c%100 == 0 {
			fmt.Printf("Completed %d queries so far...\n", c)
		}
		if c >= numQueries {
			break
		}
	}

	fmt.Printf("Completed %d queries with %d timeouts.\n", c, timeouts)
	fmt.Printf("Timing summary:\n")
	stats.PrintSummary()
	/*
		fmt.Printf("Timing percentiles:\n")
		stats.PrintPercentiles()
	*/

	//t.Fail()
}

type log struct {
	// The list of fields checked for.
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockNumber string   `json:"blockNumber"`
}

type Stats struct {
	times []time.Duration
}

func (s *Stats) Add(d time.Duration) {
	s.times = append(s.times, d)
}

func (s *Stats) PrintSummary() {
	slices.Sort(s.times)

	total := time.Duration(0)
	for _, t := range s.times {
		total += t
	}

	count := len(s.times)
	avg := total / time.Duration(count)
	p50 := s.times[count/2]
	p90 := s.times[(count*90)/100]
	p95 := s.times[(count*95)/100]
	p99 := s.times[(count*99)/100]

	fmt.Printf("Count: %d, Avg: %v, P50: %v, P90: %v, P95: %v, P99: %v\n", count, avg, p50, p90, p95, p99)
}

func (s *Stats) PrintPercentiles() {
	slices.Sort(s.times)

	count := len(s.times)
	for p := 0; p <= 100; p += 1 {
		idx := (count * p) / 100
		if idx >= count {
			idx = count - 1
		}
		fmt.Printf("P%d: %v\n", p, s.times[idx])
	}
}
