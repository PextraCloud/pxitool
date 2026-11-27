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

import "testing"

func TestExtractLVMNames(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		wantVG string
		wantLV string
		wantOk bool
	}{
		{"dev form", "/dev/myvg/mylv", "myvg", "mylv", true},
		{"dev form with dash in lv", "/dev/myvg/my-lv", "myvg", "my-lv", true},
		{"mapper form", "/dev/mapper/myvg-mylv", "myvg", "mylv", true},
		{"mapper form with escaped dash in lv", "/dev/mapper/myvg-my--lv", "myvg", "my-lv", true},
		{"mapper form with escaped dash in vg", "/dev/mapper/my--vg-mylv", "my-vg", "mylv", true},
		{"mapper form with dash in both", "/dev/mapper/my--vg-my--lv", "my-vg", "my-lv", true},
		{"mapper form with multi dash in lv", "/dev/mapper/vg-a--b--c", "vg", "a-b-c", true},
		{"no prefix", "myvg/mylv", "", "", false},
		{"empty path", "", "", "", false},
		{"dev form missing slash", "/dev/myvg", "", "", false},
		{"dev form empty vg", "/dev//mylv", "", "", false},
		{"dev form empty lv", "/dev/myvg/", "", "", false},
		{"dev form bare prefix", "/dev/", "", "", false},
		{"dev form extra path segment", "/dev/myvg/mylv/extra", "", "", false},
		{"dev form lv with trailing slash", "/dev/myvg/mylv/", "", "", false},
		{"mapper form bare prefix", "/dev/mapper/", "", "", false},
		{"mapper form missing lv separator", "/dev/mapper/myvg", "", "", false},
		{"mapper form empty lv", "/dev/mapper/myvg-", "", "", false},
		{"mapper form empty vg", "/dev/mapper/-mylv", "", "", false},
		{"mapper form trailing escape only", "/dev/mapper/myvg--", "", "", false},
		{"mapper form two single dashes", "/dev/mapper/myvg-mylv-extra", "", "", false},
		{"mapper form unescaped dash in lv", "/dev/mapper/myvg-my-lv", "", "", false},
		{"unsupported prefix", "/foo/myvg/mylv", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vg, lv, ok := ExtractLVMNames(tt.path)
			if ok != tt.wantOk {
				t.Fatalf("ExtractLVMNames(%q) ok = %v, want %v", tt.path, ok, tt.wantOk)
			}
			if vg != tt.wantVG || lv != tt.wantLV {
				t.Errorf("ExtractLVMNames(%q) = (%q, %q), want (%q, %q)", tt.path, vg, lv, tt.wantVG, tt.wantLV)
			}
		})
	}
}
