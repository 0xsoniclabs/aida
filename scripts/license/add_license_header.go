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
	copyrightYear   = "2025"
	copyrightHolder = "Sonic Labs"

	// Embedded license header content
	defaultLicenseHeader = `Copyright 2025 Sonic Labs
This file is part of Aida Testing Infrastructure for Sonic

Aida is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Aida is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with Aida. If not, see <http://www.gnu.org/licenses/>.`
)

var (
	rootDir        = flag.String("root", "", "Root directory to process (required)")
	ignorePatterns = flag.String("ignore", "carmen/,sonic/,tosca/,mock.go", "Comma-separated list of patterns to ignore")
	licenseFile    = flag.String("license-file", "", "Path to license header file (uses embedded license if not specified)")
	dryRun         = flag.Bool("dry-run", false, "Check mode: only check if license headers are correct, don't modify files")
	verbose        = flag.Bool("verbose", false, "Verbose output: show all processed files")
)

// FileStatus represents the status of a file's license header
type FileStatus int

const (
	StatusCorrect FileStatus = iota
	StatusMissing
	StatusUpdated
	StatusSkipped
)

// FileResult holds the result of processing a file
type FileResult struct {
	Path   string
	Status FileStatus
	Error  error
}

// extendLicenseHeader takes license content and prefixes each line with the comment character
func extendLicenseHeader(commentPrefix string, licenseContent string) string {
	var builder strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(licenseContent))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			builder.WriteString(commentPrefix + "\n")
		} else {
			builder.WriteString(commentPrefix + " " + line + "\n")
		}
	}

	return builder.String()
}

// readLicenseFromFile reads the license header from a file
func readLicenseFromFile(licenseFilePath string) (string, error) {
	content, err := os.ReadFile(licenseFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read license file: %w", err)
	}
	return string(content), nil
}

// shouldIgnore checks if a file path matches any ignore pattern
func shouldIgnore(path string, rootDir string, patterns []string) bool {
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

	for _, pattern := range patterns {
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

// Step 1: List all files with the given extension
func listAllFiles(rootDir string, fileExtension string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file has the correct extension
		if strings.HasSuffix(path, fileExtension) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// Step 2: Filter out files that match ignore patterns
func filterIgnoredFiles(files []string, rootDir string, patterns []string) []string {
	var filtered []string

	for _, file := range files {
		if !shouldIgnore(file, rootDir, patterns) {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// Step 3: Check which files need license header updates
func checkFilesNeedUpdate(files []string, licenseHeader string) (map[string]bool, error) {
	needsUpdate := make(map[string]bool)

	for _, file := range files {
		hasCorrectHeader, _, err := checkLicenseHeader(file, licenseHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to check %s: %w", file, err)
		}
		needsUpdate[file] = !hasCorrectHeader
	}

	return needsUpdate, nil
}

// Step 4: Apply license headers to files (or dry-run)
func applyLicenseHeaders(files []string, needsUpdate map[string]bool, licenseHeader string, commentPrefix string, dryRun bool) []FileResult {
	results := make([]FileResult, 0, len(files))

	for _, file := range files {
		result := FileResult{Path: file}

		if !needsUpdate[file] {
			result.Status = StatusCorrect
			results = append(results, result)
			continue
		}

		if dryRun {
			result.Status = StatusMissing
		} else {
			err := addLicenseToFile(file, licenseHeader, commentPrefix)
			if err != nil {
				result.Error = err
				result.Status = StatusSkipped
			} else {
				result.Status = StatusUpdated
			}
		}

		results = append(results, result)
	}

	return results
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

// processFiles orchestrates the four-step pipeline for processing files
func processFiles(rootDir string, fileExtension string, commentPrefix string, licenseHeader string, patterns []string, dryRun bool) ([]FileResult, error) {
	// Step 1: List all files
	if *verbose {
		fmt.Printf("Step 1: Listing all %s files...\n", fileExtension)
	}
	allFiles, err := listAllFiles(rootDir, fileExtension)
	if err != nil {
		return nil, err
	}
	if *verbose {
		fmt.Printf("  Found %d files\n", len(allFiles))
	}

	// Step 2: Filter ignored files
	if *verbose {
		fmt.Printf("Step 2: Filtering ignored paths...\n")
	}
	filteredFiles := filterIgnoredFiles(allFiles, rootDir, patterns)
	if *verbose {
		fmt.Printf("  %d files after filtering (%d ignored)\n", len(filteredFiles), len(allFiles)-len(filteredFiles))
	}

	// Step 3: Check which files need updates
	if *verbose {
		fmt.Printf("Step 3: Checking license headers...\n")
	}
	needsUpdate, err := checkFilesNeedUpdate(filteredFiles, licenseHeader)
	if err != nil {
		return nil, err
	}

	updateCount := 0
	for _, needs := range needsUpdate {
		if needs {
			updateCount++
		}
	}
	if *verbose {
		fmt.Printf("  %d files need updates, %d files are correct\n", updateCount, len(filteredFiles)-updateCount)
	}

	// Step 4: Apply changes (or dry-run)
	if *verbose {
		if dryRun {
			fmt.Printf("Step 4: Dry-run mode - no changes will be made\n")
		} else {
			fmt.Printf("Step 4: Applying license headers...\n")
		}
	}
	results := applyLicenseHeaders(filteredFiles, needsUpdate, licenseHeader, commentPrefix, dryRun)

	return results, nil
}

// logResults prints a summary of processing results
func logResults(results []FileResult, fileType string, dryRun bool) int {
	errorCount := 0
	correctCount := 0
	missingCount := 0
	updatedCount := 0
	skippedCount := 0

	for _, result := range results {
		relPath := result.Path
		if len(relPath) > 80 {
			relPath = "..." + relPath[len(relPath)-77:]
		}

		switch result.Status {
		case StatusCorrect:
			correctCount++
			if *verbose {
				fmt.Printf("  [OK] %s\n", relPath)
			}
		case StatusMissing:
			missingCount++
			errorCount++
			fmt.Printf("  [MISSING] %s\n", relPath)
		case StatusUpdated:
			updatedCount++
			fmt.Printf("  [UPDATED] %s\n", relPath)
		case StatusSkipped:
			skippedCount++
			errorCount++
			fmt.Printf("  [ERROR] %s: %v\n", relPath, result.Error)
		}
	}

	fmt.Printf("\nSummary for %s files:\n", fileType)
	fmt.Printf("  Total: %d\n", len(results))
	fmt.Printf("  Correct: %d\n", correctCount)
	if dryRun {
		fmt.Printf("  Missing headers: %d\n", missingCount)
	} else {
		fmt.Printf("  Updated: %d\n", updatedCount)
	}
	if skippedCount > 0 {
		fmt.Printf("  Errors: %d\n", skippedCount)
	}

	return errorCount
}

// updateCliAppCopyright updates the copyright in cli.App definitions
func updateCliAppCopyright(rootDir string, checkOnly bool) int {
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
				fmt.Printf("  [MISSING] %s (obsolete cli.App copyright)\n", path)
				errorCount++
			}
		} else {
			// Replace the copyright
			newContent := copyrightPattern.ReplaceAllString(contentStr, replacement)
			if newContent != contentStr {
				err = os.WriteFile(path, []byte(newContent), 0644)
				if err != nil {
					fmt.Printf("  [ERROR] %s: %v\n", path, err)
					errorCount++
					return nil
				}
				fmt.Printf("  [UPDATED] %s (cli.App copyright)\n", path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("  [ERROR] Failed to walk cmd directory: %v\n", err)
		errorCount++
	}

	return errorCount
}

func main() {
	flag.Parse()

	// Validate required parameter
	if *rootDir == "" {
		fmt.Fprintf(os.Stderr, "Error: --root parameter is required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s --root <directory> [options]\n\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Clean and validate the root directory
	rootDirectory := filepath.Clean(*rootDir)

	// Check if directory exists
	if _, err := os.Stat(rootDirectory); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Root directory does not exist: %s\n", rootDirectory)
		os.Exit(1)
	}

	// Parse ignore patterns from comma-separated string
	patterns := strings.Split(*ignorePatterns, ",")
	for i := range patterns {
		patterns[i] = strings.TrimSpace(patterns[i])
	}

	// Get license content
	var licenseContent string
	if *licenseFile != "" {
		// Use custom license file
		content, err := readLicenseFromFile(*licenseFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading license file: %v\n", err)
			os.Exit(1)
		}
		licenseContent = content
	} else {
		// Use embedded license
		licenseContent = defaultLicenseHeader
	}

	fmt.Printf("License Header Tool\n")
	fmt.Printf("===================\n")
	fmt.Printf("Root directory: %s\n", rootDirectory)
	if *licenseFile != "" {
		fmt.Printf("License file: %s\n", *licenseFile)
	} else {
		fmt.Printf("License: embedded (default)\n")
	}
	fmt.Printf("Ignore patterns: %s\n", strings.Join(patterns, ", "))
	if *dryRun {
		fmt.Printf("Mode: DRY-RUN (checking only, no changes will be made)\n")
	} else {
		fmt.Printf("Mode: UPDATE (files will be modified)\n")
	}
	fmt.Printf("\n")

	// Extend license header for Go files
	licenseHeader := extendLicenseHeader("//", licenseContent)

	totalErrors := 0

	// Process .go files
	fmt.Println("Processing .go files")
	fmt.Println("====================")
	results, err := processFiles(rootDirectory, ".go", "//", licenseHeader, patterns, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing .go files: %v\n", err)
		os.Exit(1)
	}
	totalErrors += logResults(results, ".go", *dryRun)

	// Update copyright in cli.App definitions
	fmt.Println("\nProcessing cli.App copyright updates")
	fmt.Println("=====================================")
	errors := updateCliAppCopyright(rootDirectory, *dryRun)
	totalErrors += errors

	// Final summary
	fmt.Printf("\n")
	fmt.Printf("===================\n")
	fmt.Printf("Final Summary\n")
	fmt.Printf("===================\n")
	if *dryRun {
		if totalErrors > 0 {
			fmt.Printf("❌ Found %d file(s) with missing or incorrect license headers\n", totalErrors)
			fmt.Printf("Run without --dry-run to fix them\n")
		} else {
			fmt.Printf("✅ All files have correct license headers\n")
		}
	} else {
		if totalErrors > 0 {
			fmt.Printf("⚠️  Completed with %d error(s)\n", totalErrors)
		} else {
			fmt.Printf("✅ All files updated successfully\n")
		}
	}

	if totalErrors > 0 {
		os.Exit(1)
	}
}
