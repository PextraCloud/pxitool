//go:build integration && linux

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
	"bytes"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

// TestBackupRBDVolume_EndToEnd creates a throwaway RBD image, backs it up, restores it to a new image, and verifies the
// restored data matches the original.
func TestBackupRBDVolume_EndToEnd(t *testing.T) {
	imageName := "pxitool2_test_image"
	runCommandFatal(t, "rbd", "create", "rbd/"+imageName, "--size", "64M")
	t.Cleanup(func() {
		exec.Command("rbd", "snap", "purge", "rbd/"+imageName).Run()
		exec.Command("rbd", "rm", "rbd/"+imageName).Run()
	})

	// Write identifiable data at a non-zero offset.
	testData := []byte("pxitool2 RBD end-to-end test data")
	tmpData, err := os.CreateTemp("", "rbd-testdata-")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpData.Write(testData); err != nil {
		t.Fatal(err)
	}
	tmpData.Close()
	defer os.Remove(tmpData.Name())
	runCommandFatal(t, "rbd", "write", "rbd/"+imageName, "--offset", "65536", "--infile", tmpData.Name())

	// Back up the image.
	workdir := t.TempDir()
	dest, err := backupRBDVolume("rbd/"+imageName, "vol1", workdir)
	if err != nil {
		t.Fatalf("backupRBDVolume() error = %v", err)
	}
	if want := path.Join(workdir, "vol1"); dest != want {
		t.Errorf("dest = %q, want %q", dest, want)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("backup stream is empty")
	}

	// Import the backup into a fresh image and read the data back.
	verifyImage := "rbd/pxitool2_test_verify"
	runCommandFatal(t, "rbd", "import", dest, verifyImage, "--no-progress")
	t.Cleanup(func() {
		exec.Command("rbd", "rm", verifyImage).Run()
	})
	readBackFile, err := os.CreateTemp("", "rbd-readback-")
	if err != nil {
		t.Fatal(err)
	}
	readBackFile.Close()
	defer os.Remove(readBackFile.Name())
	runCommandFatal(t, "rbd", "read", verifyImage, "--offset", "65536", "--length", "4096", "--outfile", readBackFile.Name())
	readBack, err := os.ReadFile(readBackFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readBack[:len(testData)], testData) {
		t.Errorf("restored data = %q, want %q", readBack[:len(testData)], testData)
	}

	// No leftover pxitool snapshots on the source image.
	out, _ := exec.Command("rbd", "snap", "ls", "rbd/"+imageName, "--format", "json").CombinedOutput()
	if strings.Contains(string(out), "pxitool-") {
		t.Errorf("leftover backup snapshot: %s", out)
	}
}
