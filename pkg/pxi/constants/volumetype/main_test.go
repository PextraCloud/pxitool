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
package volumetype

import (
	"testing"
)

func TestVolumeType_String(t *testing.T) {
	testCases := []struct {
		it       VolumeType
		expected string
	}{
		{Directory, "directory"},
		{ISCSI, "iscsi"},
		{LVM, "lvm"},
		{NetFS, "netfs"},
		{RBD, "rbd"},
		{ZFS, "zfs"},
		{LXC_, "lxc rootfs"},
	}

	for _, tc := range testCases {
		if tc.it.String() != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, tc.it.String())
		}
	}

	// Test panic on unknown type
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown VolumeType, but did not get one")
		}
	}()
	_ = (VolumeType(99)).String()
}

func TestVolumeType_MarshalText(t *testing.T) {
	testCases := []struct {
		it       VolumeType
		expected string
	}{
		{Directory, "directory"},
		{ISCSI, "iscsi"},
		{LVM, "lvm"},
		{NetFS, "netfs"},
		{RBD, "rbd"},
		{ZFS, "zfs"},
		{LXC_, "lxc rootfs"},
	}

	for _, tc := range testCases {
		data, err := tc.it.MarshalText()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if string(data) != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, string(data))
		}
	}
}

func TestVolumeType_UnmarshalText(t *testing.T) {
	testCases := []struct {
		input    string
		expected VolumeType
	}{
		{"directory", Directory},
		{"iscsi", ISCSI},
		{"lvm", LVM},
		{"netfs", NetFS},
		{"rbd", RBD},
		{"zfs", ZFS},
		{"lxc rootfs", LXC_},
	}

	for _, tc := range testCases {
		var vt VolumeType
		err := vt.UnmarshalText([]byte(tc.input))
		if err != nil {
			t.Errorf("unexpected error for input %q: %v", tc.input, err)
			continue
		}
		if vt != tc.expected {
			t.Errorf("expected %v for input %q, got %v", tc.expected, tc.input, vt)
		}
	}

	t.Run("failed unmarshal", func(t *testing.T) {
		var vt VolumeType
		err := vt.UnmarshalText([]byte("unknown"))
		if err == nil {
			t.Error("expected an error for unknown volume type, but got nil")
		}
	})
}
