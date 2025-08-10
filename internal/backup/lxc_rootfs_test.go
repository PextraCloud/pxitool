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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupLXCRootfs(t *testing.T) {
	if !commandExists("tar") {
		t.Skip("Skipping LXC test: command 'tar' not found")
	}

	// BEGIN setup temporary LXC rootfs directory
	rootDir, err := os.MkdirTemp("", "lxc-rootfs-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(rootDir)

	testData := []byte("pxitool LXC test data")
	testFile := filepath.Join(rootDir, "testfile.txt")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	// END setup temporary LXC rootfs directory

	// BEGIN test backup
	var buf bytes.Buffer
	err = BackupLXCRootfs(rootDir, &buf)
	// END test backup

	// BEGIN verify backup
	if err != nil {
		t.Errorf("BackupLXCRootfs failed: %v", err)
	} else if buf.Len() == 0 {
		t.Error("BackupLXCRootfs produced an empty backup")
	}

	// Untar archive and verify contents
	restoreDir, err := os.MkdirTemp("", "lxc-restore-test-")
	if err != nil {
		t.Fatalf("Failed to create restore dir: %v", err)
	}
	defer os.RemoveAll(restoreDir)

	cmd := exec.Command("tar", "-xf", "-", "-C", restoreDir)
	cmd.Stdin = &buf
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to untar backup stream: %v\nOutput:\n%s", err, string(output))
	}

	restoredFile := filepath.Join(restoreDir, "testfile.txt")
	restoredData, err := os.ReadFile(restoredFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if !bytes.Equal(testData, restoredData) {
		t.Errorf("Restored data does not match original data.\nOriginal: %q\nRestored: %q", testData, restoredData)
	}
	// END verify backup
}

func TestBackupLXCRootfs_FailureCases(t *testing.T) {
	if !commandExists("tar") {
		t.Skip("Skipping LXC test: command 'tar' not found")
	}

	var buf bytes.Buffer
	t.Run("non-existent directory", func(t *testing.T) {
		err := BackupLXCRootfs("/path/to/non/existent/rootfs", &buf)
		if err == nil {
			t.Error("Expected an error for a non-existent directory, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "exit status") {
			t.Errorf("Expected error to be about command exit status, but got: %v", err)
		}
	})
}
