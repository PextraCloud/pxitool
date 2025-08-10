//go:build integration && linux

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
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestBackupLVMVolume(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Skipping LVM test: must be run as root")
	}

	requiredCmds := []string{"losetup", "pvcreate", "vgcreate", "lvcreate", "dd", "lvremove", "vgremove", "pvremove"}
	for _, cmd := range requiredCmds {
		if !commandExists(cmd) {
			t.Skipf("Skipping LVM test: command '%s' not found", cmd)
		}
	}

	// BEGIN setup temporary loopback device and LVM
	backingFile, err := os.CreateTemp("", "lvm-test-backing-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	backingFileName := backingFile.Name()
	defer os.Remove(backingFileName)

	if err := backingFile.Truncate(256 * 1024 * 1024); err != nil { // 256MB
		t.Fatalf("Failed to truncate backing file: %v", err)
	}
	backingFile.Close()

	loopDeviceOutput, err := exec.Command("losetup", "-f").Output()
	if err != nil {
		t.Fatalf("Failed to find free loop device: %v", err)
	}
	loopDevice := strings.TrimSpace(string(loopDeviceOutput))
	runCommand(t, "losetup", loopDevice, backingFileName)

	vgName := "pxitool_test_vg"
	lvName := "pxitool_test_lv"
	lvPath := fmt.Sprintf("/dev/%s/%s", vgName, lvName)

	t.Cleanup(func() {
		// Ignore errors during cleanup
		exec.Command("lvremove", "-f", lvPath).Run()
		exec.Command("vgremove", "-f", vgName).Run()
		exec.Command("pvremove", "-f", loopDevice).Run()
		exec.Command("losetup", "-d", loopDevice).Run()
	})
	runCommand(t, "pvcreate", loopDevice)
	runCommand(t, "vgcreate", vgName, loopDevice)
	runCommand(t, "lvcreate", "--name", lvName, "--size", "128M", vgName)

	testData := []byte("pxitool LVM test data")
	err = os.WriteFile(lvPath, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test data to LV '%s': %v", lvPath, err)
	}
	// END setup temporary loopback device and LVM

	// BEGIN test backup
	var buf bytes.Buffer
	err = BackupLVMVolume(lvPath, &buf)
	// END test backup

	// BEGIN verify backup and data
	if err != nil {
		t.Errorf("BackupLVMVolume failed: %v", err)
	} else if buf.Len() == 0 {
		t.Error("BackupLVMVolume produced an empty backup")
	}

	backupData := make([]byte, len(testData))
	_, err = buf.Read(backupData)
	if err != nil && err != io.EOF {
		t.Fatalf("Failed to read from backup buffer: %v", err)
	}

	if !bytes.Equal(testData, backupData) {
		t.Errorf("Backup data does not match original data.\nOriginal: %q\nBackup:   %q", testData, backupData)
	}
	// END verify backup and data
}

func TestPathToLVMVolume(t *testing.T) {
	testCases := []struct {
		name       string
		path       string
		expectedVg string
		expectedLv string
	}{
		{
			name:       "valid path",
			path:       "/dev/my_vg/my_lv",
			expectedVg: "my_vg",
			expectedLv: "my_lv",
		},
		{
			name:       "path with numbers and hyphens",
			path:       "/dev/vg-01/lv-name-123",
			expectedVg: "vg-01",
			expectedLv: "lv-name-123",
		},
		{
			name:       "invalid path - no dev prefix",
			path:       "my_vg/my_lv",
			expectedVg: "",
			expectedLv: "",
		},
		{
			name:       "invalid path - too few parts",
			path:       "/dev/my_vg",
			expectedVg: "",
			expectedLv: "",
		},
		{
			name:       "invalid path - too many parts",
			path:       "/dev/my_vg/my_lv/extra",
			expectedVg: "",
			expectedLv: "",
		},
		{
			name:       "invalid path - empty string",
			path:       "",
			expectedVg: "",
			expectedLv: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vg, lv := pathToLVMVolume(tc.path)
			if vg != tc.expectedVg {
				t.Errorf("expected vg name %q, but got %q", tc.expectedVg, vg)
			}
			if lv != tc.expectedLv {
				t.Errorf("expected lv name %q, but got %q", tc.expectedLv, lv)
			}
		})
	}
}

func TestBackupLVMVolume_FailureCases(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Skipping LVM test: must be run as root")
	}
	if !commandExists("lvcreate") {
		t.Skip("Skipping LVM test: command 'lvcreate' not found")
	}

	var buf bytes.Buffer

	t.Run("non-existent volume", func(t *testing.T) {
		err := BackupLVMVolume("/dev/nonexistent_vg/nonexistent_lv", &buf)
		if err == nil {
			t.Error("Expected an error for a non-existent LVM volume, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to create LVM snapshot") {
			t.Errorf("Expected error to be about snapshot creation, but got: %v", err)
		}
	})
	t.Run("invalid volume path format", func(t *testing.T) {
		err := BackupLVMVolume("not-a-valid-path", &buf)
		if err == nil {
			t.Error("Expected an error for an invalid volume path, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to create LVM snapshot") {
			t.Errorf("Expected error to be about snapshot creation, but got: %v", err)
		}
	})
}
