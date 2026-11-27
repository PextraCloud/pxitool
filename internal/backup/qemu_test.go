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
	"strings"
	"testing"
)

func TestIsFilesystemPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"valid absolute path", "/var/lib/vm/vol1", true},
		{"root not allowed", "/", false},
		{"dot not allowed", ".", false},
		{"relative path", "relative/path", false},
		{"empty path", "", false},
		{"path too long", "/" + strings.Repeat("a", 5000), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFilesystemPath(tt.path); got != tt.want {
				t.Errorf("isFilesystemPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
