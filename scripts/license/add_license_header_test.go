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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtendLicenseHeader(t *testing.T) {
	tests := []struct {
		name           string
		commentPrefix  string
		licenseContent string
		expected       string
		shouldError    bool
	}{
		{
			name:           "Basic license with comment prefix",
			commentPrefix:  "//",
			licenseContent: "Copyright 2025\nLicense text",
			expected:       "// Copyright 2025\n// License text\n",
			shouldError:    false,
		},
		{
			name:           "License with empty lines",
			commentPrefix:  "//",
			licenseContent: "Copyright 2025\n\nLicense text",
			expected:       "// Copyright 2025\n//\n// License text\n",
			shouldError:    false,
		},
		{
			name:           "Python style comments",
			commentPrefix:  "#",
			licenseContent: "Copyright 2025",
			expected:       "# Copyright 2025\n",
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extendLicenseHeader(tt.commentPrefix, tt.licenseContent)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadLicenseFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	licenseFile := filepath.Join(tmpDir, "license.txt")
	content := "Test License\nLine 2"
	err := os.WriteFile(licenseFile, []byte(content), 0644)
	require.NoError(t, err, "Failed to create test license file")

	result, err := readLicenseFromFile(licenseFile)
	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, content, result)
}

func TestReadLicenseFromFile_NotFound(t *testing.T) {
	_, err := readLicenseFromFile("/nonexistent/file.txt")
	assert.Error(t, err, "Expected error for nonexistent file")
	assert.Contains(t, err.Error(), "failed to read license file")
}

func TestShouldIgnore(t *testing.T) {
	tmpDir := t.TempDir()
	patterns := []string{"carmen/", "sonic/", "tosca/", "mock.go"}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Carmen directory",
			path:     filepath.Join(tmpDir, "carmen", "file.go"),
			expected: true,
		},
		{
			name:     "Sonic directory",
			path:     filepath.Join(tmpDir, "sonic", "file.go"),
			expected: true,
		},
		{
			name:     "Tosca directory",
			path:     filepath.Join(tmpDir, "tosca", "file.go"),
			expected: true,
		},
		{
			name:     "Mock file",
			path:     filepath.Join(tmpDir, "test_mock.go"),
			expected: true,
		},
		{
			name:     "Hidden directory",
			path:     filepath.Join(tmpDir, ".git", "config"),
			expected: true,
		},
		{
			name:     "Regular file",
			path:     filepath.Join(tmpDir, "regular.go"),
			expected: false,
		},
		{
			name:     "Nested regular file",
			path:     filepath.Join(tmpDir, "executor", "executor.go"),
			expected: false,
		},
		{
			name:     "File with mock in middle",
			path:     filepath.Join(tmpDir, "executor", "executor_mock.go"),
			expected: true,
		},
		{
			name:     "Hidden file in root",
			path:     filepath.Join(tmpDir, ".hidden"),
			expected: true,
		},
		{
			name:     "File in tosca directory",
			path:     filepath.Join(tmpDir, "tosca", "file.go"),
			expected: true,
		},
		{
			name:     "File in sonic directory",
			path:     filepath.Join(tmpDir, "sonic", "subdir", "file.go"),
			expected: true,
		},
		{
			name:     "Deeply nested mock file",
			path:     filepath.Join(tmpDir, "a", "b", "c", "test_mock.go"),
			expected: true,
		},
		{
			name:     "File ending with mock.go pattern",
			path:     filepath.Join(tmpDir, "executor", "some_mock.go"),
			expected: true,
		},
		{
			name:     "File in carmen subdirectory",
			path:     filepath.Join(tmpDir, "carmen", "subdir", "file.go"),
			expected: true,
		},
		{
			name:     "File named exactly 'carmen'",
			path:     filepath.Join(tmpDir, "somedir", "carmen"),
			expected: false, // Files named 'carmen' are not ignored unless in a carmen/ directory
		},
		{
			name:     "File named exactly 'tosca'",
			path:     filepath.Join(tmpDir, "dir", "tosca"),
			expected: false, // Files named 'tosca' are not ignored unless in a tosca/ directory
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnore(tt.path, tmpDir, patterns)
			assert.Equal(t, tt.expected, result, "for path %s", tt.path)
		})
	}
}

func TestShouldIgnore_CurrentDir(t *testing.T) {
	tmpDir := t.TempDir()
	patterns := []string{"carmen/", "mock.go"}
	// Test with "." as part of path
	result := shouldIgnore(tmpDir, tmpDir, patterns)
	assert.False(t, result, "Current directory should not be ignored")
}

func TestShouldIgnore_InvalidPath(t *testing.T) {
	// Test with paths where relative path calculation results in .. components
	// These should be ignored since .. starts with .
	patterns := []string{"carmen/", "mock.go"}
	result := shouldIgnore("/some/path/file.go", "/completely/different/root", patterns)
	assert.True(t, result, "Paths with .. components should be ignored")
}

func TestReadFileLines(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		expected    []string
		shouldError bool
	}{
		{
			name:        "Single line",
			content:     "Line 1",
			expected:    []string{"Line 1"},
			shouldError: false,
		},
		{
			name:        "Multiple lines",
			content:     "Line 1\nLine 2\nLine 3",
			expected:    []string{"Line 1", "Line 2", "Line 3"},
			shouldError: false,
		},
		{
			name:        "Empty file",
			content:     "",
			expected:    nil, // readFileLines returns nil for empty files, which is equivalent to []string{}
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			require.NoError(t, err, "Failed to create test file")

			lines, err := readFileLines(testFile)
			if tt.shouldError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.expected, lines)
			}
		})
	}
}

func TestReadFileLines_NonExistent(t *testing.T) {
	_, err := readFileLines("/nonexistent/file.txt")
	assert.Error(t, err, "Expected error for nonexistent file")
}

func TestTrimWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"\t\nworld\n\t", "world"},
		{"no-whitespace", "no-whitespace"},
		{"   ", ""},
	}

	for _, tt := range tests {
		result := trimWhitespace(tt.input)
		assert.Equal(t, tt.expected, result, "trimWhitespace(%q)", tt.input)
	}
}

func TestCheckLicenseHeader(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name            string
		fileContent     string
		licenseHeader   string
		expectedMatch   bool
		expectedLineNum int
	}{
		{
			name:            "Correct header with blank line",
			fileContent:     "// Copyright 2025\n// License text\n\npackage main",
			licenseHeader:   "// Copyright 2025\n// License text",
			expectedMatch:   true,
			expectedLineNum: 0,
		},
		{
			name:            "Incorrect header",
			fileContent:     "// Wrong copyright\n// License text\n\npackage main",
			licenseHeader:   "// Copyright 2025\n// License text",
			expectedMatch:   false,
			expectedLineNum: 1,
		},
		{
			name:            "Missing blank line after header",
			fileContent:     "// Copyright 2025\n// License text\npackage main",
			licenseHeader:   "// Copyright 2025\n// License text",
			expectedMatch:   false,
			expectedLineNum: 3,
		},
		{
			name:            "Incomplete header",
			fileContent:     "// Copyright 2025\n",
			licenseHeader:   "// Copyright 2025\n// License text",
			expectedMatch:   false,
			expectedLineNum: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.go")
			err := os.WriteFile(testFile, []byte(tt.fileContent), 0644)
			require.NoError(t, err, "Failed to create test file")

			match, lineNum, err := checkLicenseHeader(testFile, tt.licenseHeader)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedMatch, match)
			assert.Equal(t, tt.expectedLineNum, lineNum)
		})
	}
}

func TestCheckLicenseHeader_NonExistent(t *testing.T) {
	_, _, err := checkLicenseHeader("/nonexistent/file.go", "// Header")
	assert.Error(t, err, "Expected error for nonexistent file")
}

func TestFindContentStart(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		prefix      string
		expected    int
		shouldError bool
	}{
		{
			name:        "C++ style comment block",
			content:     "/*\n * Copyright 2025\n */\n\npackage main",
			prefix:      "//",
			expected:    4,
			shouldError: false,
		},
		{
			name:        "Single line comments",
			content:     "// Copyright 2025\n// License\n\npackage main",
			prefix:      "//",
			expected:    2,
			shouldError: false,
		},
		{
			name:        "No comments",
			content:     "package main\n\nfunc main() {}",
			prefix:      "//",
			expected:    0,
			shouldError: false,
		},
		{
			name:        "All comments",
			content:     "// Line 1\n// Line 2\n// Line 3",
			prefix:      "//",
			expected:    3,
			shouldError: false,
		},
		{
			name:        "Unclosed C-style comment",
			content:     "/*\n * Copyright 2025\n * No closing",
			prefix:      "//",
			expected:    0, // First line is /*, which doesn't match //, so content starts at line 0
			shouldError: false,
		},
		{
			name:        "C-style comment with closing on same line",
			content:     "/* Copyright */\n\npackage main",
			prefix:      "//",
			expected:    0, // First line is not //, so content starts at line 0
			shouldError: false,
		},
		{
			name:        "C-style comment block with proper closing",
			content:     "/*\n * Copyright 2025\n * License\n */\n\npackage main",
			prefix:      "//",
			expected:    5, // Line 3 has */, so returns 3+2=5
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.go")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			require.NoError(t, err, "Failed to create test file")

			start, err := findContentStart(testFile, tt.prefix)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, start, "Expected start=%d, got %d", tt.expected, start)
			}
		})
	}
}

func TestFindContentStart_NonExistent(t *testing.T) {
	_, err := findContentStart("/nonexistent/file.go", "//")
	assert.Error(t, err, "Expected error for nonexistent file")
}

func TestAddLicenseToFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name            string
		initialContent  string
		licenseHeader   string
		commentPrefix   string
		expectedContent string
	}{
		{
			name:            "Add to file without header",
			initialContent:  "package main\n\nfunc main() {}",
			licenseHeader:   "// Copyright 2025\n// License",
			commentPrefix:   "//",
			expectedContent: "// Copyright 2025\n// License\npackage main\n\nfunc main() {}\n",
		},
		{
			name:            "Replace old header",
			initialContent:  "// Old copyright\n// Old license\n\npackage main",
			licenseHeader:   "// Copyright 2025\n// New license",
			commentPrefix:   "//",
			expectedContent: "// Copyright 2025\n// New license\n\npackage main\n",
		},
		{
			name:            "File with C-style comment block",
			initialContent:  "/*\n * Old copyright\n */\n\npackage main",
			licenseHeader:   "// Copyright 2025\n// New license",
			commentPrefix:   "//",
			expectedContent: "// Copyright 2025\n// New license\npackage main\n",
		},
		{
			name:            "File with only whitespace after header",
			initialContent:  "// Old\n\n\n\npackage main",
			licenseHeader:   "// New",
			commentPrefix:   "//",
			expectedContent: "// New\n\n\n\npackage main\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.go")
			err := os.WriteFile(testFile, []byte(tt.initialContent), 0644)
			require.NoError(t, err, "Failed to create test file")

			err = addLicenseToFile(testFile, tt.licenseHeader, tt.commentPrefix)
			assert.NoError(t, err, "Unexpected error")

			content, err := os.ReadFile(testFile)
			require.NoError(t, err, "Failed to read result file")

			result := string(content)
			assert.Equal(t, tt.expectedContent, result)
		})
	}
}

func TestAddLicenseToFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.go")
	err := os.WriteFile(testFile, []byte(""), 0644)
	require.NoError(t, err, "Failed to create test file")

	licenseHeader := "// Copyright 2025"
	err = addLicenseToFile(testFile, licenseHeader, "//")
	assert.NoError(t, err, "Unexpected error")

	content, _ := os.ReadFile(testFile)
	assert.True(t, strings.HasPrefix(string(content), "// Copyright 2025"), "Empty file should have license added")
}

func TestAddLicenseToFile_ReadOnlyDir(t *testing.T) {
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	os.MkdirAll(readOnlyDir, 0755)

	testFile := filepath.Join(readOnlyDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main"), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Make directory read-only
	os.Chmod(readOnlyDir, 0555)
	defer os.Chmod(readOnlyDir, 0755) // Restore permissions for cleanup

	licenseHeader := "// Copyright 2025"
	err = addLicenseToFile(testFile, licenseHeader, "//")

	// This is allowed on linux
	assert.NoError(t, err)
}

func TestProcessFiles(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "pkg"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "carmen"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755)

	files := map[string]string{
		filepath.Join(tmpDir, "test1.go"):          "package main\n\nfunc main() {}",
		filepath.Join(tmpDir, "pkg", "test2.go"):   "package pkg\n\nfunc Test() {}",
		filepath.Join(tmpDir, "carmen", "skip.go"): "package carmen", // Should be ignored
		filepath.Join(tmpDir, ".git", "config"):    "config",         // Should be ignored
		filepath.Join(tmpDir, "test.txt"):          "text file",      // Wrong extension
	}

	for path, content := range files {
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file %s", path)
	}

	licenseHeader := "// Copyright 2025\n// Test License"

	// Test with dryRun=false (should update files)
	t.Run("Update files", func(t *testing.T) {
		patterns := []string{"carmen/", "sonic/", "tosca/", "mock.go"}
		results, err := processFiles(tmpDir, ".go", "//", licenseHeader, patterns, false)
		assert.NoError(t, err, "Unexpected error")

		// Count results
		updatedCount := 0
		for _, r := range results {
			if r.Status == StatusUpdated {
				updatedCount++
			}
		}
		assert.Equal(t, 2, updatedCount, "Expected 2 files to be updated")

		// Verify files were updated
		content, _ := os.ReadFile(filepath.Join(tmpDir, "test1.go"))
		assert.True(t, strings.HasPrefix(string(content), "// Copyright 2025"), "test1.go was not updated")

		content, _ = os.ReadFile(filepath.Join(tmpDir, "pkg", "test2.go"))
		assert.True(t, strings.HasPrefix(string(content), "// Copyright 2025"), "pkg/test2.go was not updated")

		// Verify carmen directory was skipped
		content, _ = os.ReadFile(filepath.Join(tmpDir, "carmen", "skip.go"))
		assert.False(t, strings.HasPrefix(string(content), "// Copyright 2025"), "carmen/skip.go should not be updated (should be ignored)")
	})

	// Test with dryRun=true
	t.Run("Dry-run mode", func(t *testing.T) {
		tmpDir2 := t.TempDir()
		testFile := filepath.Join(tmpDir2, "test.go")
		err := os.WriteFile(testFile, []byte("package main"), 0644)
		require.NoError(t, err, "Failed to create test file")

		patterns := []string{"carmen/", "sonic/", "tosca/", "mock.go"}
		results, err := processFiles(tmpDir2, ".go", "//", licenseHeader, patterns, true)
		assert.NoError(t, err, "Unexpected error")

		// Count missing files
		missingCount := 0
		for _, r := range results {
			if r.Status == StatusMissing {
				missingCount++
			}
		}
		assert.Equal(t, 1, missingCount, "Expected 1 file to be missing header in dry-run")

		// Verify file was NOT updated
		content, _ := os.ReadFile(testFile)
		assert.False(t, strings.HasPrefix(string(content), "// Copyright 2025"), "File should not be updated in dry-run mode")
	})

	// Test with files that already have correct headers
	t.Run("Files with correct headers", func(t *testing.T) {
		tmpDir3 := t.TempDir()
		testFile := filepath.Join(tmpDir3, "correct.go")
		correctContent := licenseHeader + "\n\npackage main\n\nfunc main() {}"
		err := os.WriteFile(testFile, []byte(correctContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		patterns := []string{"carmen/", "sonic/", "tosca/", "mock.go"}
		results, err := processFiles(tmpDir3, ".go", "//", licenseHeader, patterns, false)
		assert.NoError(t, err, "Unexpected error")

		// All files should be correct
		correctCount := 0
		for _, r := range results {
			if r.Status == StatusCorrect {
				correctCount++
			}
		}
		assert.Equal(t, 1, correctCount, "Expected 1 file with correct header")

		// Verify file was not modified
		content, _ := os.ReadFile(testFile)
		assert.Equal(t, correctContent, string(content), "File with correct header should not be modified")
	})
}

func TestProcessFiles_WalkError(t *testing.T) {
	// Test with a directory that doesn't exist
	patterns := []string{"carmen/", "sonic/", "tosca/", "mock.go"}
	_, err := processFiles("/nonexistent/directory", ".go", "//", "// Header", patterns, false)
	assert.Error(t, err, "Expected error for nonexistent directory")
}

func TestProcessFiles_CheckLicenseError(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)

	// Create a file that will be unreadable to cause a read error
	testFile := filepath.Join(subDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main"), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Make the file unreadable
	os.Chmod(testFile, 0000)
	defer os.Chmod(testFile, 0644) // Restore for cleanup

	licenseHeader := "// Copyright 2025"
	patterns := []string{"carmen/", "sonic/", "tosca/", "mock.go"}
	_, err = processFiles(tmpDir, ".go", "//", licenseHeader, patterns, false)
	assert.Error(t, err)
}

func TestUpdateCliAppCopyright(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, "cmd")
	os.MkdirAll(cmdDir, 0755)

	tests := []struct {
		name         string
		content      string
		checkOnly    bool
		expectedErr  int
		shouldUpdate bool
	}{
		{
			name: "Old copyright to update",
			content: `package main
import "github.com/urfave/cli/v2"
var app = &cli.App{
	Name: "test",
	Copyright: "(c) 2022 Fantom Foundation",
}`,
			checkOnly:    false,
			expectedErr:  0,
			shouldUpdate: true,
		},
		{
			name: "Check mode with old copyright",
			content: `package main
import "github.com/urfave/cli/v2"
var app = &cli.App{
	Name: "test",
	Copyright: "(c) 2022 Fantom Foundation",
}`,
			checkOnly:    true,
			expectedErr:  1,
			shouldUpdate: false,
		},
		{
			name: "No cli.App",
			content: `package main
func main() {}`,
			checkOnly:    false,
			expectedErr:  0,
			shouldUpdate: false,
		},
		{
			name: "Already updated copyright",
			content: `package main
import "github.com/urfave/cli/v2"
var app = &cli.App{
	Name: "test",
	Copyright: "(c) 2025 Sonic Labs",
}`,
			checkOnly:    false,
			expectedErr:  0,
			shouldUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(cmdDir, "test.go")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			require.NoError(t, err, "Failed to create test file")

			errorCount := updateCliAppCopyright(tmpDir, tt.checkOnly)
			assert.Equal(t, tt.expectedErr, errorCount, "Expected %d errors, got %d", tt.expectedErr, errorCount)

			content, _ := os.ReadFile(testFile)
			contentStr := string(content)

			if tt.shouldUpdate {
				assert.Contains(t, contentStr, "Copyright: \"(c) 2025 Sonic Labs\"", "Copyright was not updated")
			} else if tt.checkOnly && strings.Contains(tt.content, "cli.App") {
				assert.NotContains(t, contentStr, "2025 Sonic Labs", "File should not be updated in check mode")
			}

			// Clean up for next test
			os.Remove(testFile)
		})
	}
}

func TestUpdateCliAppCopyright_NoCmd(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create cmd directory

	errorCount := updateCliAppCopyright(tmpDir, false)
	// Should return error when cmd directory doesn't exist
	assert.Equal(t, 1, errorCount, "Expected 1 error when cmd directory doesn't exist")
}

func TestUpdateCliAppCopyright_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, "cmd")
	os.MkdirAll(cmdDir, 0755)

	// Create multiple files with different scenarios
	files := map[string]string{
		"app1.go": `package main
import "github.com/urfave/cli/v2"
var app = &cli.App{
	Name: "app1",
	Copyright: "(c) 2022 Fantom Foundation",
}`,
		"app2.go": `package main
import "github.com/urfave/cli/v2"
var app = &cli.App{
	Name: "app2",
	Copyright: "(c) 2025 Sonic Labs",
}`,
		"no_app.go": `package main
func main() {}`,
	}

	for name, content := range files {
		path := filepath.Join(cmdDir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file %s", name)
	}

	// Test update mode
	errorCount := updateCliAppCopyright(tmpDir, false)
	assert.Equal(t, 0, errorCount, "Expected 0 errors in update mode")

	// Verify app1.go was updated
	content, _ := os.ReadFile(filepath.Join(cmdDir, "app1.go"))
	assert.Contains(t, string(content), "2025 Sonic Labs", "app1.go should have been updated")

	// Verify app2.go remained unchanged
	content, _ = os.ReadFile(filepath.Join(cmdDir, "app2.go"))
	assert.Contains(t, string(content), "2025 Sonic Labs", "app2.go should still have correct copyright")
}

func TestUpdateCliAppCopyright_ReadError(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, "cmd")
	os.MkdirAll(cmdDir, 0755)

	testFile := filepath.Join(cmdDir, "test.go")
	content := `package main
import "github.com/urfave/cli/v2"
var app = &cli.App{
	Name: "test",
	Copyright: "(c) 2022 Fantom Foundation",
}`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Make file unreadable
	os.Chmod(testFile, 0000)
	defer os.Chmod(testFile, 0644)

	errCount := updateCliAppCopyright(tmpDir, false)
	assert.Equal(t, 1, errCount, "Expected 1 error due to read failure")
}

func TestAddLicenseToFile_WithBlankLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with multiple blank lines at start
	content := "\n\n\npackage main\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err, "Failed to create test file")

	licenseHeader := "// Copyright 2025"
	err = addLicenseToFile(testFile, licenseHeader, "//")
	assert.NoError(t, err, "Unexpected error")

	result, _ := os.ReadFile(testFile)
	assert.True(t, strings.HasPrefix(string(result), "// Copyright 2025"), "License should be added before blank lines")
}

func TestAddLicenseToFile_OnlyComments(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with only comments
	content := "// Old comment\n// Another comment"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err, "Failed to create test file")

	licenseHeader := "// Copyright 2025"
	err = addLicenseToFile(testFile, licenseHeader, "//")
	assert.NoError(t, err, "Unexpected error")

	result, _ := os.ReadFile(testFile)
	expected := "// Copyright 2025\n"
	assert.Equal(t, expected, string(result))
}

func TestAddLicenseToFile_LongFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "long.go")

	// Create a file with many lines
	var builder strings.Builder
	for i := 0; i < 100; i++ {
		builder.WriteString(fmt.Sprintf("// Line %d\n", i))
	}
	builder.WriteString("\npackage main\n")

	err := os.WriteFile(testFile, []byte(builder.String()), 0644)
	require.NoError(t, err, "Failed to create test file")

	licenseHeader := "// Copyright 2025\n// License"
	err = addLicenseToFile(testFile, licenseHeader, "//")
	assert.NoError(t, err, "Unexpected error")

	result, _ := os.ReadFile(testFile)
	assert.True(t, strings.HasPrefix(string(result), "// Copyright 2025"), "License should be added to long file")
}
