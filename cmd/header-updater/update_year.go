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

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

// updateYearCommand compact given database
var updateYearCommand = cli.Command{
	Action: updateYearAction,
	Name:   "year",
	Usage:  "Increments the year in the license header of all .go files in the workspace",
}

func updateYearAction(*cli.Context) error {
	// Regex for headers like: // Copyright <year> Sonic Labs
	reHeader := regexp.MustCompile(`// Copyright (\d{4}) Sonic Labs`)

	// Regex for cli.App copyright line: Copyright: "(c) <year> Sonic Labs",
	reCLI := regexp.MustCompile(`Copyright:\s*"\(c\)\s*(\d{4})\s+Sonic Labs"`)

	excludedDirs := []string{"carmen", "tosca", "sonic"}

	return filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "mock") {
			return nil
		}

		// Skip excluded directories
		for _, ex := range excludedDirs {
			if strings.Contains(path, ex) {
				return nil
			}
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		content := string(data)
		updated := content

		// Update header years
		updated = reHeader.ReplaceAllStringFunc(updated, func(match string) string {
			matches := reHeader.FindStringSubmatch(match)
			if len(matches) != 2 {
				return match
			}
			year, _ := strconv.Atoi(matches[1])
			return fmt.Sprintf("// Copyright %d Sonic Labs", year+1)
		})

		// Update cli.App copyright
		updated = reCLI.ReplaceAllStringFunc(updated, func(match string) string {
			matches := reCLI.FindStringSubmatch(match)
			if len(matches) != 2 {
				return match
			}
			year, _ := strconv.Atoi(matches[1])
			return fmt.Sprintf(`Copyright: "(c) %d Fantom Foundation"`, year+1)
		})

		if updated != content {
			err = os.WriteFile(path, []byte(updated), 0644)
			if err != nil {
				return fmt.Errorf("failed to write %s: %w", path, err)
			}
			fmt.Println("Updated:", path)
		}

		return nil
	})
}
