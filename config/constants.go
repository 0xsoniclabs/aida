package config

import (
	"github.com/0xsoniclabs/aida/config/chainid"
	"math"
)

const (
	Unknown chainid.EthTestType = iota
	StateTests
	BlockTests
)
const (
	AidaDbRepositorySonicUrl    = "https://storage.googleapis.com/aida-repository-public/sonic/aida-patches"
	AidaDbRepositoryOperaUrl    = "https://storage.googleapis.com/aida-repository-public/mainnet/aida-patches"
	AidaDbRepositoryTestnetUrl  = "https://storage.googleapis.com/aida-repository-public/testnet/aida-patches"
	AidaDbRepositoryEthereumUrl = "https://storage.googleapis.com/aida-repository-public/ethereum/aida-patches"
	AidaDbRepositoryHoleskyUrl  = "https://storage.googleapis.com/aida-repository-public/holesky/aida-patches"
	AidaDbRepositoryHoodiUrl    = "https://storage.googleapis.com/aida-repository-public/hoodi/aida-patches"
	AidaDbRepositorySepoliaUrl  = "https://storage.googleapis.com/aida-repository-public/sepolia/aida-patches"
)

const maxLastBlock = math.MaxUint64 - 1 // we decrease the value by one because params are always +1

// A map of key blocks on Fantom chain
var KeywordBlocks = map[chainid.ChainID]map[string]uint64{
	chainid.SonicMainnetChainID: {
		"zero":        0,
		"opera":       0,
		"istanbul":    0,
		"muirglacier": 0,
		"berlin":      0,
		"london":      0,
		"shanghai":    0, //timestamp
		"cancun":      0, //timestamp
		"first":       0,
		"last":        maxLastBlock,
		"lastpatch":   0,
	},
	chainid.MainnetChainID: {
		"zero":        0,
		"opera":       4_564_026,
		"istanbul":    0, // todo istanbul block for mainnet?
		"muirglacier": 0, // todo muirglacier block for mainnet?
		"berlin":      37_455_223,
		"london":      37_534_833,
		"shanghai":    maxLastBlock, //timestamp
		"cancun":      maxLastBlock, //timestamp
		"first":       0,
		"last":        maxLastBlock,
		"lastpatch":   0,
	},
	chainid.TestnetChainID: {
		"zero":        0,
		"opera":       479_327,
		"istanbul":    0, // todo istanbul block for testnet?
		"muirglacier": 0, // todo muirglacier block for testnet?
		"berlin":      1_559_470,
		"london":      7_513_335,
		"shanghai":    maxLastBlock, //timestamp
		"cancun":      maxLastBlock, //timestamp
		"first":       0,
		"last":        maxLastBlock,
		"lastpatch":   0,
	},
	// ethereum fork blocks are not stored in this structure as ethereum has already prepared config
	// at params.MainnetChainConfig and it has bigger amount of forks than Fantom chain
	// Ethereum config - https://github.com/ethereum/go-ethereum/blob/3e4fbce034b384c99afeead6cf0f72be0a2b8f13/params/config.go#L43
	chainid.EthereumChainID: {
		"zero":        0,
		"opera":       0,
		"istanbul":    9_069_000,
		"muirglacier": 9_200_000,
		"berlin":      12_244_000,
		"london":      12_965_000,
		"shanghai":    1681338455, //timestamp
		"cancun":      1710338135, //timestamp
		"first":       0,
		"last":        maxLastBlock,
		"lastpatch":   0,
	},
	// Holesky config - https://github.com/ethereum/go-ethereum/blob/7d8aca95d28c4e8560a657fd1ff7852ad4eee72c/params/config.go#L69C2-L69C36
	chainid.HoleskyChainID: {
		"zero":        0,
		"opera":       0,
		"istanbul":    0,
		"muirglacier": maxLastBlock, // is nil in geth implementation, probably all functionality is overwritten by later forks
		"berlin":      0,
		"london":      0,
		"shanghai":    1696000704, //timestamp
		"cancun":      1707305664, //timestamp
		"prague":      1740434112, //timestamp
		"first":       0,
		"last":        maxLastBlock,
		"lastpatch":   0,
	},
	// Hoodi config - https://github.com/ethereum/go-ethereum/blob/3e4fbce034b384c99afeead6cf0f72be0a2b8f13/params/config.go#L130
	chainid.HoodiChainID: {
		"zero":        0,
		"opera":       0,
		"istanbul":    0,
		"muirglacier": 0,
		"berlin":      0,
		"london":      0,
		"shanghai":    0,          //timestamp
		"cancun":      0,          //timestamp
		"prague":      1742999832, //timestamp
		"first":       0,
		"last":        maxLastBlock,
		"lastpatch":   0,
	},
	// Sepolia config - https://github.com/ethereum/go-ethereum/blob/3e4fbce034b384c99afeead6cf0f72be0a2b8f13/params/config.go#L100
	chainid.SepoliaChainID: {
		"zero":        0,
		"opera":       0,
		"istanbul":    0,
		"muirglacier": 0,
		"berlin":      0,
		"london":      0,
		"shanghai":    1677557088, //timestamp
		"cancun":      1706655072, //timestamp
		"prague":      1741159776, //timestamp
		"first":       0,
		"last":        maxLastBlock,
		"lastpatch":   0,
	},

	// EthTest must always set its fork blocks to 0 because each test has random block number
	// and if that block number is not greater than the config, the test won't get executed
	chainid.EthTestsChainID: {},
}

const (
	// special transaction number for pseudo transactions
	PseudoTx = 99999
)

// GitCommit represents the GitHub commit hash the app was built from.
var GitCommit = "0000000000000000000000000000000000000000"
