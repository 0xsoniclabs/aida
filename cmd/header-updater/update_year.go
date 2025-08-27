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
