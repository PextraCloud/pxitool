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
package names

import (
	"strings"
	"testing"
)

func TestExtractZFSName(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		want   string
		wantOk bool
	}{
		{"device path", "/dev/zvol/pool/dataset", "pool/dataset", true},
		{"bare form", "pool/dataset", "pool/dataset", true},
		{"nested dataset", "/dev/zvol/pool/dataset/sub", "pool/dataset/sub", true},
		{"bare nested dataset", "pool/dataset/sub", "pool/dataset/sub", true},
		{"max length", "/dev/zvol/" + strings.Repeat("p", 100) + "/" + strings.Repeat("d", 153), strings.Repeat("p", 100) + "/" + strings.Repeat("d", 153), true},
		{"over max length", "/dev/zvol/" + strings.Repeat("p", 100) + "/" + strings.Repeat("d", 154), "", false},
		{"over max length bare", strings.Repeat("p", 100) + "/" + strings.Repeat("d", 154), "", false},
		{"empty path", "", "", false},
		{"bare prefix only", "/dev/zvol/", "", false},
		{"missing dataset", "/dev/zvol/pool", "", false},
		{"empty dataset", "/dev/zvol/pool/", "", false},
		{"empty pool", "/dev/zvol//dataset", "", false},
		{"no separator", "pool", "", false},
		{"trailing slash", "pool/", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ok := ExtractZFSName(tt.path)
			if ok != tt.wantOk {
				t.Fatalf("ExtractZFSName(%q) ok = %v, want %v", tt.path, ok, tt.wantOk)
			}
			if name != tt.want {
				t.Errorf("ExtractZFSName(%q) = %q, want %q", tt.path, name, tt.want)
			}
		})
	}
}
