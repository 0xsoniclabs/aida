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

package coverage

import (
	"fmt"
	"strings"
	"sync"
)

const carmenModulePrefix = "github.com/0xsoniclabs/carmen/go"

var (
	coverPkgOnce     sync.Once
	coverPkgWarn     sync.Once
	coverPkgPrefixes []string
	corePrefixes     = []string{
		"github.com/0xsoniclabs/carmen/go/carmen",
		"github.com/0xsoniclabs/carmen/go/state",
		"github.com/0xsoniclabs/carmen/go/database/mpt",
	}
)

// isValidCarmenPackage returns true when a package path belongs to the Carmen module.
// Only packages within github.com/0xsoniclabs/carmen/go count towards coverage.
func isValidCarmenPackage(pkgPath string) bool {
	if pkgPath == "" {
		return false
	}
	if pkgPath == carmenModulePrefix {
		return true
	}
	return strings.HasPrefix(pkgPath, carmenModulePrefix+"/")
}

// filterCarmenUnits keeps only coverage units that belong to the Carmen module.
func filterCarmenUnits(units map[CounterKey]unitInfo) map[CounterKey]unitInfo {
	allowedPrefixes := getCoverPkgPrefixes()
	filtered := make(map[CounterKey]unitInfo, len(units))
	for key, info := range units {
		if !isValidCarmenPackage(info.PackagePath) {
			continue
		}
		if len(allowedPrefixes) > 0 && !hasAllowedPrefix(info.PackagePath, allowedPrefixes) {
			continue
		}
		filtered[key] = info
	}

	// If .coverpkgs filtering removed everything, fall back to all Carmen packages
	if len(filtered) == 0 && len(allowedPrefixes) > 0 {
		filtered = make(map[CounterKey]unitInfo, len(units))
		for key, info := range units {
			if isValidCarmenPackage(info.PackagePath) {
				filtered[key] = info
			}
		}
		if len(filtered) > 0 {
			coverPkgWarn.Do(func() {
				fmt.Printf("coverage: filter removed all Carmen packages; falling back to full Carmen module\n")
			})
		}
	}
	return filtered
}

// filterCountersForUnits keeps only counters that correspond to the provided units.
// Missing counters are initialised to zero so downstream logic can rely on presence.
func filterCountersForUnits(counts map[CounterKey]uint32, units map[CounterKey]unitInfo) map[CounterKey]uint32 {
	filtered := make(map[CounterKey]uint32, len(units))
	for key := range units {
		filtered[key] = 0
	}
	for key, value := range counts {
		if _, ok := units[key]; ok {
			filtered[key] = value
		}
	}
	return filtered
}

func hasAllowedPrefix(pkg string, allowed []string) bool {
	for _, prefix := range allowed {
		if strings.HasPrefix(pkg, prefix) {
			return true
		}
	}
	return false
}

func getCoverPkgPrefixes() []string {
	coverPkgOnce.Do(func() {
		coverPkgPrefixes = corePrefixes
	})
	return coverPkgPrefixes
}
