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
		Arg(path.Join(testDataDir, "events.json")).
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
