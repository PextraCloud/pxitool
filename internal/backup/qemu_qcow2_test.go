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
	"strings"
	"testing"
)

func TestBackupQEMUVolume(t *testing.T) {
	if !commandExists("qemu-img") {
		t.Skip("Skipping QEMU test: command 'qemu-img' not found")
	}

	// BEGIN setup source qcow2 file
	sourceFile, err := os.CreateTemp("", "pxitool-test-source-*.qcow2")
	if err != nil {
		t.Fatalf("Failed to create temp source file: %v", err)
	}
	sourceFileName := sourceFile.Name()
	sourceFile.Close()
	defer os.Remove(sourceFileName)

	runCommand(t, "qemu-img", "create", "-f", "qcow2", sourceFileName, "10M")
	// END setup source qcow2 file

	// BEGIN test backup
	var buf bytes.Buffer
	err = BackupQEMUVolume(sourceFileName, &buf)
	// END test backup

	// BEGIN verify backup
	if err != nil {
		t.Errorf("BackupQEMUVolume failed: %v", err)
	} else if buf.Len() == 0 {
		t.Error("BackupQEMUVolume produced an empty backup")
	}

	// TODO: write some test data and verify it

	// Verify that the output is a valid qcow2 file
	verifyFile, err := os.CreateTemp("", "pxitool-test-verify-*.qcow2")
	if err != nil {
		t.Fatalf("Failed to create temp verify file: %v", err)
	}
	verifyFileName := verifyFile.Name()
	defer os.Remove(verifyFileName)

	if _, err := verifyFile.Write(buf.Bytes()); err != nil {
		t.Fatalf("Failed to write buffer to verify file: %v", err)
	}
	verifyFile.Close()

	runCommand(t, "qemu-img", "info", verifyFileName)
	// END verify backup
}

func TestBackupQEMUVolume_FailureCases(t *testing.T) {
	if !commandExists("qemu-img") {
		t.Skip("Skipping QEMU test: command 'qemu-img' not found")
	}

	var buf bytes.Buffer

	t.Run("non-existent file", func(t *testing.T) {
		err := BackupQEMUVolume("/path/to/non/existent/file.qcow2", &buf)
		if err == nil {
			t.Error("Expected an error for a non-existent file, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to open source file") {
			t.Errorf("Expected error to be about opening source file, but got: %v", err)
		}
	})
}
