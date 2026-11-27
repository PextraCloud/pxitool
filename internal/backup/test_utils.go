/*
Copyright 2026 Pextra Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package backup

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

// runCommandFatal runs a command and fails the test if it exits non-zero.
func runCommandFatal(t *testing.T, name string, args ...string) {
	t.Helper()
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
}

// firstDiff returns the index of the first byte that differs between a and b, or -1 if they are equal.
func firstDiff(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	if len(a) != len(b) {
		return n
	}
	return -1
}

// firstNonPattern returns the index of the first byte != pattern, or -1.
func firstNonPattern(data []byte, pattern byte) int {
	for i, b := range data {
		if b != pattern {
			return i
		}
	}
	return -1
}

// dirNames returns the names of the entries in a directory.
func dirNames(entries []os.DirEntry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name()
	}
	return out
}

// tarList returns the member paths of the archive at path (via `tar -tf`). All paths have no trailing slashes.
func tarList(t *testing.T, path string) []string {
	t.Helper()
	out, err := exec.Command("tar", "-tf", path).Output()
	if err != nil {
		t.Fatalf("tar -tf %s: %v", path, err)
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			names = append(names, strings.TrimSuffix(line, "/"))
		}
	}
	return names
}

// tarExtract returns the bytes of the archive member path (as listed by tar) extracted from the archive at archivePath
// into a fresh directory.
func tarExtract(t *testing.T, archivePath, member string) []byte {
	t.Helper()
	dest := t.TempDir()
	if out, err := exec.Command("tar", "-xf", archivePath, "-C", dest, member).CombinedOutput(); err != nil {
		t.Fatalf("tar -xf %s %s: %v\n%s", archivePath, member, err, out)
	}
	b, err := os.ReadFile(path.Join(dest, strings.TrimPrefix(member, "./")))
	if err != nil {
		t.Fatalf("read extracted %s: %v", member, err)
	}
	return b
}
