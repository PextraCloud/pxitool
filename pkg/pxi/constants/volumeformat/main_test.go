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
package volumeformat

import (
	"testing"
)

func TestVolumeFormat_String(t *testing.T) {
	testCases := []struct {
		it       VolumeFormat
		expected string
	}{
		{Raw, "raw"},
		{QCOW2, "qcow2"},
		{VMDK, "vmdk"},
	}

	for _, tc := range testCases {
		if tc.it.String() != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, tc.it.String())
		}
	}

	// Test panic on unknown type
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown VolumeFormat, but did not get one")
		}
	}()
	_ = (VolumeFormat(99)).String()
}

func TestVolumeFormat_MarshalText(t *testing.T) {
	testCases := []struct {
		it       VolumeFormat
		expected string
	}{
		{Raw, "raw"},
		{QCOW2, "qcow2"},
		{VMDK, "vmdk"},
	}

	for _, tc := range testCases {
		data, err := tc.it.MarshalText()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if string(data) != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, data)
		}
	}
}
func TestVolumeFormat_UnmarshalText(t *testing.T) {
	testCases := []struct {
		input    string
		expected VolumeFormat
	}{
		{"raw", Raw},
		{"qcow2", QCOW2},
		{"vmdk", VMDK},
	}

	for _, tc := range testCases {
		var vf VolumeFormat
		err := vf.UnmarshalText([]byte(tc.input))
		if err != nil {
			t.Errorf("unexpected error for input %q: %v", tc.input, err)
			continue
		}
		if vf != tc.expected {
			t.Errorf("expected %v, got %v for input %q", tc.expected, vf, tc.input)
		}
	}

	var vf VolumeFormat
	err := vf.UnmarshalText([]byte("unknown"))
	if err == nil {
		t.Error("expected error for unknown VolumeFormat, but did not get one")
	}
}
