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
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	licenseFile     = "license_header.txt"
	copyrightYear   = "2025"
	copyrightHolder = "Sonic Labs"
)

var (
	// List of file/directory patterns to ignore
	ignorePatterns = []string{
		"carmen/",
		"sonic/",
		"tosca/",
		"mock.go",
	}

	checkOnly = flag.Bool("check", false, "Only check if license headers are correct, don't modify files")
)

// extendLicenseHeader reads the license header file and prefixes each line with the comment character
func extendLicenseHeader(commentPrefix string, licenseFilePath string) (string, error) {
	file, err := os.Open(licenseFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open license file: %w", err)
	}
	defer file.Close()

	var builder strings.Builder
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			builder.WriteString(commentPrefix + "\n")
		} else {
			builder.WriteString(commentPrefix + " " + line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read license file: %w", err)
	}

	return builder.String(), nil
}

// shouldIgnore checks if a file path matches any ignore pattern
func shouldIgnore(path string, rootDir string) bool {
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return false
	}

	// Check if any directory component starts with "."
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." {
			return true
		}
	}

	for _, pattern := range ignorePatterns {
		if strings.Contains(relPath, pattern) {
			return true
		}
		if strings.HasSuffix(relPath, pattern) {
			return true
		}
	}
	return false
}

// readFileLines reads a file and returns its lines
func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// trimWhitespace trims leading and trailing whitespace from a string
func trimWhitespace(s string) string {
	return strings.TrimSpace(s)
}

// checkLicenseHeader checks if the file has the correct license header
func checkLicenseHeader(filePath string, licenseHeader string) (bool, int, error) {
	fileLines, err := readFileLines(filePath)
	if err != nil {
		return false, 0, err
	}

	licenseLines := strings.Split(strings.TrimRight(licenseHeader, "\n"), "\n")

	// Check if each line of the license header matches the file
	for i, licenseLine := range licenseLines {
		if i >= len(fileLines) {
			return false, i + 1, nil
		}

		if trimWhitespace(fileLines[i]) != trimWhitespace(licenseLine) {
			return false, i + 1, nil
		}
	}

	// Check if the line after the license header is empty or whitespace-only
	lineAfterHeader := len(licenseLines)
	if lineAfterHeader < len(fileLines) {
		if trimWhitespace(fileLines[lineAfterHeader]) != "" {
			return false, lineAfterHeader + 1, nil
		}
	}

	return true, 0, nil
}

// findContentStart finds the line number where the actual content starts (after old header)
func findContentStart(filePath string, commentPrefix string) (int, error) {
	fileLines, err := readFileLines(filePath)
	if err != nil {
		return 0, err
	}

	commentPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(commentPrefix))

	// Find first line that doesn't match the comment prefix
	for i, line := range fileLines {
		if !commentPattern.MatchString(line) {
			// Check for C++ style comments /* */
			if i == 0 && strings.TrimSpace(line) == "/*" {
				// Find the closing */
				for j := i + 1; j < len(fileLines); j++ {
					if strings.Contains(fileLines[j], "*/") {
						return j + 2, nil // +2 to skip the */ line and the blank line after
					}
				}
			}
			return i, nil
		}
	}

	return len(fileLines), nil
}

// addLicenseToFile adds or updates the license header in a file
func addLicenseToFile(filePath string, licenseHeader string, commentPrefix string) error {
	// Find where the actual content starts
	startLine, err := findContentStart(filePath, commentPrefix)
	if err != nil {
		return err
	}

	fileLines, err := readFileLines(filePath)
	if err != nil {
		return err
	}

	// Build new content
	var builder strings.Builder
	builder.WriteString(licenseHeader)
	builder.WriteString("\n") // Add blank line after license header

	// Append the rest of the file content
	if startLine < len(fileLines) {
		for i := startLine; i < len(fileLines); i++ {
			builder.WriteString(fileLines[i] + "\n")
		}
	}

	// Write the new content
	err = os.WriteFile(filePath, []byte(builder.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// addLicenseToFiles processes all files with the given extension
func addLicenseToFiles(rootDir string, fileExtension string, commentPrefix string, licenseHeader string, checkOnly bool) (int, error) {
	errorCount := 0
	fileCount := 0

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Print current directory being processed
		if info.IsDir() {
			relPath, _ := filepath.Rel(rootDir, path)
			if relPath == "." {
				relPath = "root"
			}

			// Check if this directory should be skipped
			if shouldIgnore(path, rootDir) {
				// fmt.Printf("  [SKIP] %s\n", relPath)
				return filepath.SkipDir
			}

			// fmt.Printf("  [SCAN] %s\n", relPath)
			return nil
		}

		// Check if file should be ignored
		if shouldIgnore(path, rootDir) {
			return nil
		}

		// Check if file has the correct extension
		if !strings.HasSuffix(path, fileExtension) {
			return nil
		}

		fileCount++

		// Check if license header is correct
		hasCorrectHeader, _, err := checkLicenseHeader(path, licenseHeader)
		if err != nil {
			return fmt.Errorf("failed to check license header in %s: %w", path, err)
		}

		if !hasCorrectHeader {
			if checkOnly {
				fmt.Printf("  [MISSING] %s\n", path)
				errorCount++
			} else {
				err := addLicenseToFile(path, licenseHeader, commentPrefix)
				if err != nil {
					return fmt.Errorf("failed to add license to %s: %w", path, err)
				}
				fmt.Printf("  [UPDATED] %s\n", path)
			}
		}

		return nil
	})

	if err != nil {
		return errorCount, err
	}

	fmt.Printf("\nScanned %d files with extension %s\n", fileCount, fileExtension)
	return errorCount, nil
}

// updateCliAppCopyright updates the copyright in cli.App definitions
func updateCliAppCopyright(rootDir string, checkOnly bool) (int, error) {
	errorCount := 0
	cmdDir := filepath.Join(rootDir, "cmd")

	copyrightPattern := regexp.MustCompile(`Copyright:\s*"[^"]*"`)
	replacement := fmt.Sprintf(`Copyright: "(c) %s %s"`, copyrightYear, copyrightHolder)
	oldCopyright := `Copyright: "(c) 2022 Fantom Foundation"`

	err := filepath.Walk(cmdDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		contentStr := string(content)

		// Check if file contains cli.App
		if !strings.Contains(contentStr, "cli.App{") {
			return nil
		}

		if checkOnly {
			if strings.Contains(contentStr, oldCopyright) {
				fmt.Printf("Obsolete cli.App copyright in: %s\n", path)
				errorCount++
			}
		} else {
			// Replace the copyright
			newContent := copyrightPattern.ReplaceAllString(contentStr, replacement)
			err = os.WriteFile(path, []byte(newContent), 0644)
			if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return errorCount, err
	}

	return errorCount, nil
}

func main() {
	flag.Parse()

	// Get the script directory
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}
	scriptDir := filepath.Dir(execPath)

	// Resolve the root directory (two levels up from script directory)
	rootDir := filepath.Clean(filepath.Join(scriptDir, "..", ".."))

	// Check if we're in the scripts/license directory (when running via `go run`)
	if strings.HasSuffix(scriptDir, "scripts/license") || strings.HasSuffix(scriptDir, "scripts\\license") {
		rootDir = filepath.Clean(filepath.Join(scriptDir, "..", ".."))
	}

	licenseFilePath := filepath.Join(scriptDir, licenseFile)

	// Extend license header for Go files
	licenseHeader, err := extendLicenseHeader("//", licenseFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading license header: %v\n", err)
		os.Exit(1)
	}

	totalErrors := 0

	// Update .go files
	fmt.Println("Updating .go files")
	errors, err := addLicenseToFiles(rootDir, ".go", "//", licenseHeader, *checkOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing .go files: %v\n", err)
		os.Exit(1)
	}
	totalErrors += errors

	// Update copyright in cli.App definitions
	fmt.Println("Updating copyright in cli.App definitions")
	errors, err = updateCliAppCopyright(rootDir, *checkOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating cli.App copyright: %v\n", err)
		os.Exit(1)
	}
	totalErrors += errors

	if totalErrors > 0 {
		os.Exit(1)
	}
}
