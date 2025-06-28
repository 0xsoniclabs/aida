package utildb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type DumpAccount struct {
	Balance  string `json:"balance,omitempty"`
	Nonce    uint64 `json:"nonce,omitempty"`
	CodeHash string `json:"codeHash,omitempty"`
	Address  string `json:"address,omitempty"`
	Key      string `json:"key,omitempty"`
	Root     string `json:"root,omitempty"`
}

// LoadEthereumDumpWorldState loads world state from ethereum dump file.
func LoadEthereumDumpWorldState(dump string) (chan *DumpAccount, chan error, error) {
	fmt.Println("Loading Ethereum dump world state from", dump)
	accountChan := make(chan *DumpAccount, 100)
	errChan := make(chan error, 1)

	go func() {
		file, err := os.Open(dump)
		if err != nil {
			errChan <- fmt.Errorf("closing accountChan due to error: %v", err)
			close(accountChan)
			close(errChan)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var acc DumpAccount
			if err := json.Unmarshal([]byte(line), &acc); err != nil {
				continue
			}
			accountChan <- &acc
		}
		if err := scanner.Err(); err != nil {
			errChan <- err
		}
		close(accountChan)
		close(errChan)
	}()

	return accountChan, errChan, nil
}
