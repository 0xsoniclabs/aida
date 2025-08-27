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
	return updateYear(".")
}

// updateYear walks through files and updates copyright years.
// Returns an error if something goes wrong.
func updateYear(root string) error {
	reHeader := regexp.MustCompile(`// Copyright (\d{4}) Sonic Labs`)
	reCLI := regexp.MustCompile(`Copyright:\s*"\(c\)\s*(\d{4})\s+Fantom Foundation"`)

	excludedDirs := []string{"carmen", "tosca", "sonic"}

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "mock") {
			return nil
		}
		for _, ex := range excludedDirs {
			if strings.Contains(path, string(filepath.Separator)+ex+string(filepath.Separator)) {
				return nil
			}
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		updated := content

		updated = reHeader.ReplaceAllStringFunc(updated, func(match string) string {
			matches := reHeader.FindStringSubmatch(match)
			year, _ := strconv.Atoi(matches[1])
			return fmt.Sprintf("// Copyright %d Sonic Labs", year+1)
		})

		updated = reCLI.ReplaceAllStringFunc(updated, func(match string) string {
			matches := reCLI.FindStringSubmatch(match)
			year, _ := strconv.Atoi(matches[1])
			return fmt.Sprintf(`Copyright: "(c) %d Fantom Foundation"`, year+1)
		})

		if updated != content {
			return os.WriteFile(path, []byte(updated), 0644)
		}
		return nil
	})
}
