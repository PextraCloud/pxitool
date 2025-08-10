/*
Copyright 2025 Pextra Inc.

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
	"os/exec"
	"strings"
	"testing"
)

// commandExists checks if a command is available in the system's PATH.
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// runCommand executes a command and fails the test if it errors.
func runCommand(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command '%s %s' failed: %v\nOutput:\n%s", name, strings.Join(args, " "), err, string(output))
	}
}
