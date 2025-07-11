// Copyright 2024 Fantom Foundation
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

package utils

import (
	crytporand "crypto/rand"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
)

const testAccountStorageSize = 10

type StateDbTestCase struct {
	Variant        string
	ShadowImpl     string
	archiveMode    bool
	ArchiveVariant string
	primeRandom    bool
}

func GetStateDbTestCases() []StateDbTestCase {
	testCases := []StateDbTestCase{
		{"geth", "", true, "", false},
		{"geth", "geth", true, "", false},
		{"carmen", "geth", false, "none", false},
		{"carmen", "geth", true, "ldb", false},
		{"carmen", "geth", true, "sqlite", false},
	}

	return testCases
}

// MakeRandomByteSlice creates byte slice of given length with randomized values
func MakeRandomByteSlice(t *testing.T, bufferLength int) []byte {
	// make byte slice
	buffer := make([]byte, bufferLength)

	// fill the slice with random data
	_, err := crytporand.Read(buffer)
	if err != nil {
		t.Fatalf("failed test data; can not generate random byte slice; %s", err.Error())
	}

	return buffer
}

// GetRandom generates random number in from given range
func GetRandom(rangeLower int, rangeUpper int) int {
	// seed the PRNG
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// get randomized balance
	randInt := rangeLower + rand.Intn(rangeUpper-rangeLower+1)
	return randInt
}

// MakeAccountStorage generates randomized account storage with testAccountStorageSize length
func MakeAccountStorage(t *testing.T) map[common.Hash]common.Hash {
	// create storage map
	storage := map[common.Hash]common.Hash{}

	// fill the storage map
	for j := 0; j < testAccountStorageSize; j++ {
		k := common.BytesToHash(MakeRandomByteSlice(t, 32))
		storage[k] = common.BytesToHash(MakeRandomByteSlice(t, 32))
	}

	return storage
}

// MakeTestConfig creates a config struct for testing
func MakeTestConfig(testCase StateDbTestCase) *Config {
	cfg := &Config{
		DbLogging:      "",
		DbImpl:         testCase.Variant,
		DbVariant:      "",
		ShadowImpl:     testCase.ShadowImpl,
		ShadowVariant:  "",
		ArchiveVariant: testCase.ArchiveVariant,
		ArchiveMode:    testCase.archiveMode,
		PrimeRandom:    testCase.primeRandom,
		ChainID:        MainnetChainID,
	}

	if testCase.Variant == "flat" {
		cfg.DbVariant = "go-memory"
	}

	if testCase.primeRandom {
		cfg.PrimeThreshold = 0
		cfg.RandomSeed = int64(GetRandom(1_000_000, 100_000_000))
	}

	return cfg
}

// MakeWorldState generates randomized world state containing 100 accounts
func MakeWorldState(t *testing.T) (txcontext.AidaWorldState, []common.Address) {
	// create list of addresses
	var addrList []common.Address

	// create world state
	ws := make(map[common.Address]txcontext.Account)

	for i := 0; i < 100; i++ {
		// create random address
		addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

		// add to address list
		addrList = append(addrList, addr)

		acc := txcontext.NewAccount(
			MakeRandomByteSlice(t, 2048),
			MakeAccountStorage(t),
			big.NewInt(int64(GetRandom(1, 1000*5000))),
			uint64(GetRandom(1, 1000*5000)),
		)
		ws[addr] = acc

		// create account

	}

	return ws, addrList
}
