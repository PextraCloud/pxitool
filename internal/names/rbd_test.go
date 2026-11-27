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

func TestExtractRBDNames(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantPool  string
		wantImage string
		wantOk    bool
	}{
		{"simple", "pool/image", "pool", "image", true},
		{"single char names", "p/i", "p", "i", true},
		{"max length", strings.Repeat("p", 128) + "/" + strings.Repeat("i", 126), strings.Repeat("p", 128), strings.Repeat("i", 126), true},
		{"starts with slash", "/pool/image", "", "", false},
		{"empty path", "", "", "", false},
		{"no slash", "pool", "", "", false},
		{"empty image", "pool/", "", "", false},
		{"empty pool", "/image", "", "", false},
		{"image with extra slash", "pool/image/extra", "", "", false},
		{"over max length", strings.Repeat("p", 128) + "/" + strings.Repeat("i", 127), "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, image, ok := ExtractRBDNames(tt.path)
			if ok != tt.wantOk {
				t.Fatalf("ExtractRBDNames(%q) ok = %v, want %v", tt.path, ok, tt.wantOk)
			}
			if pool != tt.wantPool || image != tt.wantImage {
				t.Errorf("ExtractRBDNames(%q) = (%q, %q), want (%q, %q)", tt.path, pool, image, tt.wantPool, tt.wantImage)
			}
		})
	}
}
