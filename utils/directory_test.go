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

package utils

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFilesWithinDirectories(t *testing.T) {
	tmpA := t.TempDir()
	tmpB := t.TempDir()

	// Prepare all tested files
	allFiles := []string{filepath.Join(tmpA, "fileA"), filepath.Join(tmpB, "fileB"), filepath.Join(tmpB, "File.suf")}
	for _, fName := range allFiles {
		f, err := os.Create(fName)
		if err != nil {
			t.Fatal(err)
		}
		err = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name   string
		want   []string
		suffix string
	}{
		{
			name: "No-suffix_Expect-all",
			want: allFiles,
		},
		{
			name:   "Has-suffix_Expect-one",
			want:   []string{filepath.Join(tmpB, "File.suf")},
			suffix: "suf",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			got, err := GetFilesWithinDirectories(test.suffix, []string{tmpA, tmpB})
			if err != nil {
				t.Fatal(err)
			}

			slices.Sort(got)
			slices.Sort(test.want)
			if slices.Compare(got, test.want) != 0 {
				t.Errorf("unexpected files\n got: %v\nwant: %v", got, test.want)
			}
		})
	}
}

func TestGetFreeSpace(t *testing.T) {
	// Valid directory should return a positive free space
	dir := t.TempDir()
	space, err := GetFreeSpace(dir)
	assert.NoError(t, err)
	assert.Greater(t, space, int64(0))

	// Non-existent directory should return an error
	invalid := filepath.Join(dir, "does_not_exist")
	_, err = GetFreeSpace(invalid)
	assert.Error(t, err)
}

func TestGetDirectorySize(t *testing.T) {
	// Setup a temporary directory with files
	root := t.TempDir()
	file1 := filepath.Join(root, "file1.txt")
	content1 := []byte("hello")
	err := os.WriteFile(file1, content1, 0644)
	assert.NoError(t, err)

	subdir := filepath.Join(root, "sub")
	err = os.Mkdir(subdir, 0755)
	assert.NoError(t, err)

	file2 := filepath.Join(subdir, "file2.txt")
	content2 := []byte("world!")
	err = os.WriteFile(file2, content2, 0644)
	assert.NoError(t, err)

	// Verify total size equals sum of file sizes
	size, err := GetDirectorySize(root)
	assert.NoError(t, err)
	expected := int64(len(content1) + len(content2))
	assert.Equal(t, expected, size)

	// Non-existent directory should return an nil
	invalid := filepath.Join(root, "no_such_dir")
	_, err = GetDirectorySize(invalid)
	assert.Nil(t, err)
}
