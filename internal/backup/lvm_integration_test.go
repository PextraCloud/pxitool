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
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"testing"
)

// lvmTestVolume sets up a throwaway file-backed PV/VG/LV and registers cleanup. Returns the LV device path.
func lvmTestVolume(t *testing.T, vgName, lvName string) string {
	t.Helper()
	backingFile, err := os.CreateTemp("", "lvm-test-backing-")
	if err != nil {
		t.Fatalf("creating backing file: %v", err)
	}
	if err := backingFile.Truncate(256 * 1024 * 1024); err != nil {
		backingFile.Close()
		os.Remove(backingFile.Name())
		t.Fatalf("truncating backing file: %v", err)
	}
	backingFile.Close()

	loopOut, err := exec.Command("losetup", "-f").Output()
	if err != nil {
		t.Fatalf("finding free loop device: %v", err)
	}
	loopDevice := strings.TrimSpace(string(loopOut))

	runCommandFatal(t, "losetup", loopDevice, backingFile.Name())
	t.Cleanup(func() {
		exec.Command("losetup", "-d", loopDevice).Run()
		exec.Command("pvremove", "-f", loopDevice).Run()
		exec.Command("vgremove", "-f", vgName).Run()
		os.Remove(backingFile.Name())
	})

	runCommandFatal(t, "pvcreate", loopDevice)
	runCommandFatal(t, "vgcreate", vgName, loopDevice)
	runCommandFatal(t, "lvcreate", "--name", lvName, "--size", "128M", vgName)
	return fmt.Sprintf("/dev/%s/%s", vgName, lvName)
}

// TestBackupLVMVolume_EndToEnd creates a throwaway PV/VG/LV, backs it up, and verifies the exported bytes match the data written to the LV.
func TestBackupLVMVolume_EndToEnd(t *testing.T) {
	lvPath := lvmTestVolume(t, "pxitool2_test_vg", "pxitool2_test_lv")
	testData := []byte("pxitool2 LVM end-to-end test data")
	if err := os.WriteFile(lvPath, testData, 0644); err != nil {
		t.Fatalf("writing test data to %s: %v", lvPath, err)
	}

	workdir := t.TempDir()
	dest, err := backupLVMVolume(lvPath, "vol1", workdir)
	if err != nil {
		t.Fatalf("backupLVMVolume() error = %v", err)
	}
	if want := path.Join(workdir, "vol1"); dest != want {
		t.Errorf("dest = %q, want %q", dest, want)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("backup is empty")
	}
	// The LV is 128M; the backup must be a full-size image.
	if len(data) != 128*1024*1024 {
		t.Errorf("backup len = %d, want 128MiB", len(data))
	}
	if !bytes.Equal(data[:len(testData)], testData) {
		t.Errorf("restored data = %q, want %q", data[:len(testData)], testData)
	}
}

// TestBackupLVMVolume_LiveConcurrentWriters tests a live backup scenario. A writer continuously rewrites a 4K-aligned
// block of the LV with a deterministic pattern while the backup runs. The test verifies that the backup completes, the
// exported image is full-size and self-consistent, and no pxitool- snapshot is left behind.
func TestBackupLVMVolume_LiveConcurrentWriters(t *testing.T) {
	const (
		blockSize = 4096
		block     = 16 // 4K-aligned offset: 16 blocks
		numWrites = 2
	)

	lvPath := lvmTestVolume(t, "pxitool2_test_vg_live", "pxitool2_test_lv_live")

	// Deterministic pattern; the writer rewrites whole 4K-aligned blocks with this identical pattern.
	pattern := make([]byte, blockSize)
	for i := range pattern {
		pattern[i] = byte(i % 251)
	}

	stop := make(chan struct{})
	var wg sync.WaitGroup

	// Writers: continuously rewrite the payload block with the same pattern.
	for w := 0; w < numWrites; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f, err := os.OpenFile(lvPath, os.O_WRONLY, 0)
			if err != nil {
				t.Errorf("writer: open: %v", err)
				return
			}
			defer f.Close()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := f.WriteAt(pattern, int64(block*blockSize)); err != nil {
					t.Errorf("writer: write: %v", err)
					return
				}
			}
		}()
	}

	workdir := t.TempDir()
	dest, err := backupLVMVolume(lvPath, "vol1", workdir)
	close(stop)
	wg.Wait()
	if err != nil {
		t.Fatalf("backupLVMVolume() under live I/O: %v", err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 128*1024*1024 {
		t.Fatalf("backup len = %d, want 128MiB (truncated backup)", len(data))
	}

	// Verify the pattern
	got := data[block*blockSize : (block+1)*blockSize]
	if i := firstDiff(got, pattern); i >= 0 {
		t.Errorf("payload block is not a coherent pattern (first mismatch at %d)", i)
	}

	// Verify no leftover pxitool- snapshot exists
	out, _ := exec.Command("lvs", "--noheadings", "-o", "name", "pxitool2_test_vg_live").CombinedOutput()
	if strings.Contains(string(out), "pxitool-") {
		t.Errorf("leftover backup snapshot: %s", out)
	}
}
