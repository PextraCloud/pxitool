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
package version

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestVersion_DarUnavailable(t *testing.T) {
	// Ensure dar is not available in PATH
	t.Setenv("PATH", "")

	got := Version()
	want := fmt.Sprintf("pxitool %s (unknown (error retrieving dar version))", PxiToolVersion)
	if got != want {
		t.Errorf("Version() = %q, want %q", got, want)
	}
}

func TestVersion_DarAvailable(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/sh\nprintf '  dar 3.0.0\\nlibdar 2.7.2\\n'\n"
	if err := os.WriteFile(filepath.Join(dir, "dar"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	got := Version()
	want := fmt.Sprintf("pxitool %s (dar 3.0.0, libdar 2.7.2)", PxiToolVersion)
	if got != want {
		t.Errorf("Version() = %q, want %q", got, want)
	}
}
