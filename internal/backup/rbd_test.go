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
	"strings"
	"testing"
)

// checkCephReady skips the test if a Ceph environment is not detected.
func checkCephReady(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Skipping RBD test: must be run as root")
	}
	if !commandExists("rbd") {
		t.Skip("Skipping RBD test: command 'rbd' not found")
	}
	// A ceph.conf is a good indicator that a cluster is configured.
	if _, err := os.Stat("/etc/ceph/ceph.conf"); os.IsNotExist(err) {
		t.Skip("Skipping RBD test: /etc/ceph/ceph.conf not found")
	}
}

func TestBackupRBDVolume(t *testing.T) {
	checkCephReady(t)

	// Assume a Ceph pool named 'rbd' exists.
	poolName := "rbd"
	imageName := "pxitool_test_image"
	rbdSpec := fmt.Sprintf("%s/%s", poolName, imageName)

	// BEGIN setup RBD image
	runCommand(t, "rbd", "create", rbdSpec, "--size", "128M")

	t.Cleanup(func() {
		// Ignore errors during cleanup
		exec.Command("rbd", "snap", "purge", rbdSpec).Run()
		exec.Command("rbd", "rm", rbdSpec).Run()
	})
	// END setup RBD image

	// BEGIN write test data
	testData := []byte("pxitool RBD test data")
	offset := "0"
	tmpFile, err := os.CreateTemp("", "pxitool_testdata")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()
	runCommand(t, "rbd", "write", rbdSpec, "--offset", offset, "--infile", tmpFile.Name())
	// END write test data

	// BEGIN test backup
	var buf bytes.Buffer
	err = BackupRBDVolume(rbdSpec, &buf)
	// END test backup

	// BEGIN verify backup
	if err != nil {
		t.Errorf("BackupRBDVolume failed: %v", err)
	} else if buf.Len() == 0 {
		t.Error("BackupRBDVolume produced an empty backup")
	}

	// Extract the backup stream to a temp file for verification
	backupFile, err := os.CreateTemp("", "pxitool_backup")
	if err != nil {
		t.Fatalf("Failed to create temp backup file: %v", err)
	}
	defer os.Remove(backupFile.Name())
	if _, err := backupFile.Write(buf.Bytes()); err != nil {
		t.Fatalf("Failed to write backup data: %v", err)
	}
	backupFile.Close()

	// Import backup to a new image for verification
	verifyImage := fmt.Sprintf("%s/%s_verify", poolName, imageName)
	runCommand(t, "rbd", "import", backupFile.Name(), verifyImage, "--no-progress")
	defer exec.Command("rbd", "rm", verifyImage).Run()

	// Read back the data from the verify image
	readFile, err := os.CreateTemp("", "pxitool_readback")
	if err != nil {
		t.Fatalf("Failed to create temp readback file: %v", err)
	}
	defer os.Remove(readFile.Name())
	runCommand(t, "rbd", "read", verifyImage, "--offset", offset, "--length", fmt.Sprintf("%d", len(testData)), "--outfile", readFile.Name())
	readBytes, err := os.ReadFile(readFile.Name())
	if err != nil {
		t.Fatalf("Failed to read back data: %v", err)
	}
	if !bytes.Equal(readBytes, testData) {
		t.Errorf("Data mismatch: expected %q, got %q", testData, readBytes)
	}
	// END verify backup
}

func TestBackupRBDVolume_FailureCases(t *testing.T) {
	checkCephReady(t)

	var buf bytes.Buffer

	t.Run("non-existent image", func(t *testing.T) {
		err := BackupRBDVolume("rbd/nonexistent_image_for_pxitool", &buf)
		if err == nil {
			t.Error("Expected an error for a non-existent RBD image, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to create RBD snapshot") {
			t.Errorf("Expected error to be about snapshot creation, but got: %v", err)
		}
	})
}
