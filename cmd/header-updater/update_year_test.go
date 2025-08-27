package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper to write a test file and return its path
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	return path
}

// helper to read file contents
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(data)
}

func TestUpdateCopyrights_HeaderUpdate(t *testing.T) {
	tmp := t.TempDir()
	path := writeFile(t, tmp, "file.go", `// Copyright 2025 Sonic Labs
package main
`)

	err := updateYear(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "// Copyright 2026 Sonic Labs") {
		t.Errorf("expected updated year, got:\n%s", got)
	}
}

func TestUpdateCopyrights_CLIUpdate(t *testing.T) {
	tmp := t.TempDir()
	path := writeFile(t, tmp, "cli.go", `
package main

var RunVMApp = cli.App{
	Name:      "Aida",
	Copyright: "(c) 2023 Fantom Foundation",
}
`)

	err := updateYear(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "(c) 2024 Fantom Foundation") {
		t.Errorf("expected updated year, got:\n%s", got)
	}
}

func TestUpdateCopyrights_NoChange(t *testing.T) {
	tmp := t.TempDir()
	path := writeFile(t, tmp, "other.go", `package main`)

	err := updateYear(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if got != "package main" {
		t.Errorf("expected no change, got:\n%s", got)
	}
}

func TestUpdateCopyrights_ExcludedFiles(t *testing.T) {
	tmp := t.TempDir()

	// mock file should be ignored
	mockPath := writeFile(t, tmp, "mock_test.go", `// Copyright 2025 Sonic Labs`)

	// file under carmen dir should be ignored
	carmenDir := filepath.Join(tmp, "carmen")
	os.Mkdir(carmenDir, 0755)
	carmenPath := writeFile(t, carmenDir, "c.go", `// Copyright 2025 Sonic Labs`)

	err := updateYear(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ensure contents are unchanged
	if got := readFile(t, mockPath); !strings.Contains(got, "2025") {
		t.Errorf("mock file should not be updated")
	}
	if got := readFile(t, carmenPath); !strings.Contains(got, "2025") {
		t.Errorf("carmen file should not be updated")
	}
}
