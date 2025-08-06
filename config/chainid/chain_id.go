package chainid

import "github.com/urfave/cli/v2"

// ChainID is typed int64 same as in go-ethereum
type ChainID int64
type ChainIDs map[ChainID]string
type EthTestType int

// An enums of argument modes used by trace subcommands

const (
	UnknownChainID      ChainID = 0
	EthereumChainID     ChainID = 1
	SonicMainnetChainID ChainID = 146
	MainnetChainID      ChainID = 250
	TestnetChainID      ChainID = 4002
	HoleskyChainID      ChainID = 17000
	HoodiChainID        ChainID = 560048
	SepoliaChainID      ChainID = 11155111
	// EthTestsChainID is a mock ChainID which is necessary for setting
	// the chain rules to allow any block number for any fork.
	EthTestsChainID ChainID = 1337
)

var RealChainIDs = ChainIDs{
	SonicMainnetChainID: "mainnet-sonic",
	MainnetChainID:      "mainnet-opera",
	TestnetChainID:      "testnet",
	EthereumChainID:     "ethereum",
	HoleskyChainID:      "holesky",
	HoodiChainID:        "hoodi",
	SepoliaChainID:      "sepolia",
}
var AllowedChainIDs = ChainIDs{
	SonicMainnetChainID: "mainnet-sonic",
	MainnetChainID:      "mainnet-opera",
	TestnetChainID:      "testnet",
	EthereumChainID:     "ethereum",
	HoleskyChainID:      "holesky",
	HoodiChainID:        "hoodi",
	SepoliaChainID:      "sepolia",
	EthTestsChainID:     "eth-tests",
}
var EthereumChainIDs = ChainIDs{
	EthereumChainID: "ethereum",
	HoleskyChainID:  "holesky",
	HoodiChainID:    "hoodi",
	SepoliaChainID:  "sepolia",
	EthTestsChainID: "eth-tests",
}

var ChainIDFlag = cli.IntFlag{
	Name:  "chainid",
	Usage: "ChainID for replayer",
}

// IsEthereumNetwork checks if the ChainID is an Ethereum network - mainnet, holesky, hoodi or sepolia.
// Special conditions for miner rewards and validation are applied.
func IsEthereumNetwork(chainID ChainID) bool {
	_, ok := EthereumChainIDs[chainID]
	return ok
}
