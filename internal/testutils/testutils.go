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
package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// writeStub writes a stub script with the given name and content to a temporary directory and makes it executable.
func writeStub(t *testing.T, dir, name, script string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}

// prependPath prepends the given directory to the PATH environment variable for the duration of the test.
func prependPath(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// StubScript creates a stub script with the given name and content, and prepends its directory to the PATH.
func StubScript(t *testing.T, name, script string) {
	t.Helper()
	dir := t.TempDir()
	writeStub(t, dir, name, script)
	prependPath(t, dir)
}

// StubExit creates a stub script that exits with the specified exit code, and prepends its directory to the PATH.
func StubExit(t *testing.T, name string, exitCode int) {
	t.Helper()
	dir := t.TempDir()
	writeStub(t, dir, name, fmt.Sprintf("#!/bin/sh\nexit %d\n", exitCode))
	prependPath(t, dir)
}

// StubQemuImg creates a stub for the qemu-img command that copies the source file to the destination, and prepends its
// directory to the PATH.
func StubQemuImg(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	writeStub(t, dir, "qemu-img", "#!/bin/sh\ncp \"$7\" \"$8\"\n")
	prependPath(t, dir)
}

// StubDarLog creates a stub for the dar command that logs its arguments to the specified log file, and prepends its
// directory to the PATH.
func StubDarLog(t *testing.T, dir, logFile string) {
	t.Helper()
	writeStub(t, dir, "dar", fmt.Sprintf(`#!/bin/sh
echo "$@" > "%s"
out=""
prev=""
for arg in "$@"; do
  if [ "$prev" = "-c" ]; then out="$arg"; fi
  prev="$arg"
done
[ -n "$out" ] && touch "$out"
`, logFile))
	prependPath(t, dir)
}
