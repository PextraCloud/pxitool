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
package pxiversion

import (
	"testing"
)

func TestPXIVersion_String(t *testing.T) {
	testCases := []struct {
		it       PXIVersion
		expected string
	}{
		{V1, "PXI Version 1.0"},
	}

	for _, tc := range testCases {
		if tc.it.String() != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, tc.it.String())
		}
	}

	// Test panic on unknown type
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown PXIVersion, but did not get one")
		}
	}()
	_ = (PXIVersion(99)).String()
}
