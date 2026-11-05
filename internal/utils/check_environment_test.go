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
package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeToolDir creates a temporary directory containing an executable file named
// toolName, suitable for use as a PATH entry.
func makeToolDir(t *testing.T, toolName string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, toolName)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("failed to create fake %s: %v", toolName, err)
	}
	return dir
}

func TestCheckEnvironment(t *testing.T) {
	const missingTool = "definitely-not-a-real-tool-pxitool-test"
	tests := []struct {
		name       string
		pathTools  []string // tools to place in the fake PATH
		tools      []string // tools to check
		wantErr    bool
		wantErrSub string // substring expected in the error, if any
	}{
		{
			name:      "empty tool list",
			pathTools: nil,
			tools:     nil,
			wantErr:   false,
		},
		{
			name:      "all tools found",
			pathTools: []string{"tool-a", "tool-b"},
			tools:     []string{"tool-a", "tool-b"},
			wantErr:   false,
		},
		{
			name:       "single tool missing",
			pathTools:  nil,
			tools:      []string{missingTool},
			wantErr:    true,
			wantErrSub: missingTool + " not found in PATH",
		},
		{
			name:       "one of many missing",
			pathTools:  []string{"tool-a"},
			tools:      []string{"tool-a", missingTool},
			wantErr:    true,
			wantErrSub: missingTool + " not found in PATH",
		},
		{
			name:       "multiple tools missing",
			pathTools:  nil,
			tools:      []string{missingTool + "-1", missingTool + "-2"},
			wantErr:    true,
			wantErrSub: "environment check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pathParts []string
			for _, tool := range tt.pathTools {
				pathParts = append(pathParts, makeToolDir(t, tool))
			}
			if len(pathParts) == 0 {
				// Empty PATH ensures only the fake tools are discoverable.
				pathParts = []string{t.TempDir()}
			}
			t.Setenv("PATH", strings.Join(pathParts, string(os.PathListSeparator)))

			err := CheckEnvironment(tt.tools)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("CheckEnvironment() = nil, want error")
				}
				if tt.wantErrSub != "" && !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Errorf("CheckEnvironment() error = %q, want it to contain %q", err, tt.wantErrSub)
				}
			} else {
				if err != nil {
					t.Fatalf("CheckEnvironment() = %v, want nil", err)
				}
			}
		})
	}
}
