//go:build integration

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
	"bytes"
	"testing"

	"github.com/PextraCloud/pxitool/pkg/pxi/constants/volumetype"
)

func TestBackupVolume_Failures(t *testing.T) {
	// TODO: Add non-failure test cases for BackupVolume
	testCases := []struct {
		name       string
		volumeType volumetype.VolumeType
		volumePath string
	}{
		{
			name:       "Directory type with non-existent path",
			volumeType: volumetype.Directory,
			volumePath: "/path/to/non/existent/file.qcow2",
		},
		{
			name:       "LVM type with non-existent volume",
			volumeType: volumetype.LVM,
			volumePath: "/dev/nonexistent_vg/nonexistent_lv",
		},
		{
			name:       "ZFS type with non-existent dataset",
			volumeType: volumetype.ZFS,
			volumePath: "nonexistent_pool/nonexistent_fs",
		},
		{
			name:       "RBD type with non-existent image",
			volumeType: volumetype.RBD,
			volumePath: "rbd/nonexistent_image_for_pxitool",
		},
		{
			name:       "LXC type with non-existent path",
			volumeType: volumetype.LXC_,
			volumePath: "/path/to/non/existent/rootfs",
		},
		{
			name:       "ISCSI type should always fail",
			volumeType: volumetype.ISCSI,
			volumePath: "iqn.2025-01.com.pextra:target",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip tests for tools that are not installed
			switch tc.volumeType {
			case volumetype.Directory, volumetype.NetFS:
				if !commandExists("qemu-img") {
					t.Skip("Skipping test: qemu-img not found")
				}
			case volumetype.LVM:
				if !commandExists("lvcreate") {
					t.Skip("Skipping test: lvcreate not found")
				}
			case volumetype.ZFS:
				if !commandExists("zfs") {
					t.Skip("Skipping test: zfs not found")
				}
			case volumetype.RBD:
				if !commandExists("rbd") {
					t.Skip("Skipping test: rbd not found")
				}
			case volumetype.LXC_:
				if !commandExists("tar") {
					t.Skip("Skipping test: tar not found")
				}
			}

			var buf bytes.Buffer
			_, _, err := BackupVolume(tc.volumePath, tc.volumeType, &buf)
			if err == nil {
				t.Fatalf("Expected an error for volume type %s, but got nil", tc.volumeType)
			}
		})
	}
}
