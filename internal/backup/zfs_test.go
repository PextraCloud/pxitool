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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupZFSVolume(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Skipping ZFS test: must be run as root")
	}

	requiredCmds := []string{"truncate", "zpool", "zfs"}
	for _, cmd := range requiredCmds {
		if !commandExists(cmd) {
			t.Skipf("Skipping ZFS test: command '%s' not found", cmd)
		}
	}

	// BEGIN setup temporary backing file and ZFS pool/dataset
	backingFile, err := os.CreateTemp("", "zfs-test-backing-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	backingFileName := backingFile.Name()
	defer os.Remove(backingFileName)

	if err := backingFile.Truncate(256 * 1024 * 1024); err != nil { // 256MB
		t.Fatalf("Failed to truncate backing file: %v", err)
	}
	backingFile.Close()

	poolName := "pxitool_test_pool"
	datasetName := fmt.Sprintf("%s/pxitool_test_fs", poolName)
	mountPoint := "/" + datasetName

	t.Cleanup(func() {
		// Ignore errors during cleanup
		exec.Command("zfs", "destroy", "-r", datasetName).Run()
		exec.Command("zpool", "destroy", "-f", poolName).Run()
	})
	runCommand(t, "zpool", "create", "-f", poolName, backingFileName)
	runCommand(t, "zfs", "create", datasetName)

	testData := []byte("pxitool ZFS test data")
	testFilePath := filepath.Join(mountPoint, "testfile")
	err = os.WriteFile(testFilePath, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test data to ZFS dataset '%s': %v", datasetName, err)
	}
	// END setup temporary backing file and ZFS pool/dataset

	// BEGIN test backup
	var buf bytes.Buffer
	err = BackupZFSVolume(datasetName, &buf)
	// END test backup

	// BEGIN verify backup and data using zfs recv
	if err != nil {
		t.Errorf("BackupZFSVolume failed: %v", err)
	} else if buf.Len() == 0 {
		t.Error("BackupZFSVolume produced an empty backup")
	}

	// Restore the backup into a new dataset using zfs recv
	recvDataset := fmt.Sprintf("%s_restored", datasetName)
	t.Cleanup(func() {
		exec.Command("zfs", "destroy", "-r", recvDataset).Run()
	})

	cmd := exec.Command("zfs", "recv", "-F", recvDataset)
	cmd.Stdin = &buf
	var recvOut bytes.Buffer
	cmd.Stdout = &recvOut
	cmd.Stderr = &recvOut
	if err := cmd.Run(); err != nil {
		t.Fatalf("zfs recv failed: %v, output: %s", err, recvOut.String())
	}

	// Check that the restored file exists and contents match
	restoredMount := "/" + recvDataset
	restoredFile := filepath.Join(restoredMount, "testfile")
	restoredData, err := os.ReadFile(restoredFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if !bytes.Equal(restoredData, testData) {
		t.Errorf("Restored file contents do not match original. Got: %q, want: %q", restoredData, testData)
	}
	// END verify backup and data using zfs recv
}

func TestPathToZFSDataset(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "valid zvol path",
			path:     "/dev/zvol/my_pool/my_dataset",
			expected: "my_pool/my_dataset",
		},
		{
			name:     "path with numbers and hyphens",
			path:     "/dev/zvol/pool-01/data-set-123",
			expected: "pool-01/data-set-123",
		},
		{
			name:     "path without zvol prefix",
			path:     "my_pool/my_dataset",
			expected: "my_pool/my_dataset",
		},
		{
			name:     "invalid path - not a zvol",
			path:     "/dev/sda1",
			expected: "/dev/sda1",
		},
		{
			name:     "invalid path - too short",
			path:     "/dev/zvol",
			expected: "/dev/zvol",
		},
		{
			name:     "invalid path - empty string",
			path:     "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pathToZFSDataset(tc.path)
			if result != tc.expected {
				t.Errorf("expected dataset %q, but got %q", tc.expected, result)
			}
		})
	}
}

func TestBackupZFSVolume_FailureCases(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Skipping ZFS test: must be run as root")
	}
	if !commandExists("zfs") {
		t.Skip("Skipping ZFS test: command 'zfs' not found")
	}

	var buf bytes.Buffer

	t.Run("non-existent dataset", func(t *testing.T) {
		err := BackupZFSVolume("nonexistent_pool/nonexistent_fs", &buf)
		if err == nil {
			t.Error("Expected an error for a non-existent ZFS dataset, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to create snapshot") {
			t.Errorf("Expected error to be about snapshot creation, but got: %v", err)
		}
	})
}
