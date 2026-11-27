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
	"os"
	"path"
	"strings"
	"testing"

	"github.com/PextraCloud/pxitool/pkg/pxi"
)

func TestBackupVolume(t *testing.T) {
	// For "lxc rootfs not a dir" case
	lxcNotDirCase := path.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(lxcNotDirCase, []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}
	// For "qemu not a file" (directory) case
	qemuNotAFile := path.Join(t.TempDir(), "not-a-file")
	if err := os.Mkdir(qemuNotAFile, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		volumeType pxi.VolumeType
		sourcePath string
		wantErr    string
	}{
		{"directory", pxi.VolumeTypeDirectory, "/pxitool-test-nonexistent-file", "failed to open source file"},
		{"directory invalid path", pxi.VolumeTypeDirectory, "not-a-valid-path", "source path is not a valid filesystem path"},
		{"directory not a file", pxi.VolumeTypeDirectory, qemuNotAFile, "is a directory"},
		{"lvm", pxi.VolumeTypeLVM, "not-a-valid-path", "unsupported LVM volume path"},
		{"zfs", pxi.VolumeTypeZFS, "not-a-valid-path", "invalid ZFS zvol path"},
		{"rbd", pxi.VolumeTypeRBD, "not-a-valid-path", "unsupported RBD image path"},
		{"lxc rootfs", pxi.VolumeTypeLXCRootfs, "/pxitool-test-nonexistent-dir", "failed to stat source"},
		{"lxc rootfs invalid path", pxi.VolumeTypeLXCRootfs, "not-a-valid-path", "source path is not a valid filesystem path"},
		{"lxc rootfs not a dir", pxi.VolumeTypeLXCRootfs, lxcNotDirCase, "is not a directory"},
		{"unsupported volume type", pxi.VolumeType("foo"), "/pxitool-test-nonexistent-file", "unsupported volume type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BackupVolume(tt.sourcePath, "vol1", tt.volumeType, t.TempDir())
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
