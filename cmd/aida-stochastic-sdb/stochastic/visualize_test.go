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

package stochastic

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunStochasticVisualizeCommand(t *testing.T) {
	// given
	app := cli.NewApp()
	app.Commands = []*cli.Command{&StochasticVisualizeCommand}
	port := "8182"
	args := utils.NewArgs("test").
		Arg(StochasticVisualizeCommand.Name).
		Flag(utils.PortFlag.Name, port).
		Arg(path.Join(testDataDir, "stats.json")).
		Build()

	// create a context with timeout to prevent the test from hanging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// channel to communicate errors from the goroutine
	errChan := make(chan error, 1)

	// start the web server in a goroutine since app.Run is blocking
	go func() {
		err := app.Run(args)
		errChan <- err
	}()

	// wait for the server to start up
	serverURL := fmt.Sprintf("http://localhost:%s", port)

	// try to connect to the server with retries
	var resp *http.Response
	var err error
	maxRetries := 10
	retryDelay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			t.Fatal("Test timeout reached while waiting for server to start")
		case err := <-errChan:
			if err != nil {
				t.Fatalf("Server failed to start: %v", err)
			}
		default:
			client := &http.Client{Timeout: 2 * time.Second}
			resp, err = client.Get(serverURL)
			if err == nil {
				break
			}
			time.Sleep(retryDelay)
		}
	}

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, resp.Body.Close())
}
