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
package dar

import (
	"bytes"
	"fmt"
	"os/exec"
)

// cmdError formats the error message for a failed command execution, including the exit code and stderr output.
func cmdError(err error, stderr bytes.Buffer) error {
	if err == nil {
		return nil
	}

	processState, ok := err.(*exec.ExitError)
	if !ok {
		return fmt.Errorf("command failed: %w", err)
	}
	return fmt.Errorf("command failed with exit code %d: %s", processState.ExitCode(), stderr.String())
}
