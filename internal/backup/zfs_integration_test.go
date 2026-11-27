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

	"github.com/PextraCloud/pxitool/pkg/pxi"
)

// zvolTestPool creates a throwaway file-backed pool with a zvol of the given size and registers cleanup. The caller
// writes through /dev/zvol/<zvol>.
func zvolTestPool(t *testing.T, poolName, zvolName, size string) {
	t.Helper()
	backingFile, err := os.CreateTemp("", "zfs-test-backing-")
	if err != nil {
		t.Fatalf("creating backing file: %v", err)
	}
	if err := backingFile.Truncate(256 * 1024 * 1024); err != nil {
		backingFile.Close()
		os.Remove(backingFile.Name())
		t.Fatalf("truncating backing file: %v", err)
	}
	backingFile.Close()

	runCommandFatal(t, "zpool", "create", "-f", poolName, backingFile.Name())
	t.Cleanup(func() {
		exec.Command("zfs", "destroy", "-R", zvolName).Run()
		exec.Command("zpool", "destroy", "-f", poolName).Run()
		os.Remove(backingFile.Name())
	})
	runCommandFatal(t, "zfs", "create", "-V", size, zvolName)

	runCommandFatal(t, "udevadm", "settle", "--timeout=5")
}

// assertNoPxitoolSnapshots fails the test if any pxitool- tagged snapshot is still attached to the dataset (i.e. the
// backup did not clean up).
func assertNoPxitoolSnapshots(t *testing.T, dataset string) {
	t.Helper()
	out, _ := exec.Command("zfs", "list", "-H", "-o", "name", "-t", "snapshot", "-r", dataset).CombinedOutput()
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.Contains(line, "pxitool-") {
			t.Errorf("leftover backup snapshot: %q", line)
		}
	}
}

// writeZvolPattern fills the 4k-aligned block at bs=4k seek=block of the zvol with one copy of pattern (4k-aligned
// via dd).
func writeZvolPattern(t *testing.T, dev string, pattern []byte, block int) {
	t.Helper()
	tmp, err := os.CreateTemp("", "zfs-pattern-")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.Write(pattern); err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	runCommandFatal(t, "dd", "if="+tmp.Name(), "of="+dev, "bs=4k",
		fmt.Sprintf("seek=%d", block), "conv=notrunc", "status=none")
}

// ddReadBlock reads one 4k-aligned block from dev via dd.
func ddReadBlock(t *testing.T, dev string, block int) []byte {
	t.Helper()
	out, err := os.CreateTemp("", "zfs-readback-")
	if err != nil {
		t.Fatal(err)
	}
	out.Close()
	defer os.Remove(out.Name())
	cmd := exec.Command("dd", "if="+dev, "of="+out.Name(), "bs=4k",
		fmt.Sprintf("skip=%d", block), "count=1", "status=none")
	if o, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("dd read %s block %d: %v\n%s", dev, block, err, o)
	}
	b, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBackupZFSVolume_EndToEnd(t *testing.T) {
	poolName := "pxitool2_test_pool"
	zvolName := poolName + "/zvol_test"

	zvolTestPool(t, poolName, zvolName, "64M")
	testData := []byte("pxitool2 ZFS end-to-end test data")

	// Write the test data to the zvol at 4k-aligned offset 16 (64k).
	writeZvolPattern(t, "/dev/zvol/"+zvolName, testData, 16)

	// Back up the zvol (device path form).
	workdir := t.TempDir()
	dest, err := backupZFSVolume("/dev/zvol/"+zvolName, "vol1", workdir)
	if err != nil {
		t.Fatalf("backupZFSVolume() error = %v", err)
	}
	if want := path.Join(workdir, "vol1"); dest != want {
		t.Errorf("dest = %q, want %q", dest, want)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("destination missing: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("backup stream is empty")
	}

	// Expect the backup stream to recover into a new dataset with the same contents.
	recvDataset := poolName + "/zvol_test_recv"
	recvCmd := exec.Command("zfs", "recv", "-F", recvDataset)
	recvCmd.Stdin = bytes.NewReader(data)
	if out, err := recvCmd.CombinedOutput(); err != nil {
		t.Fatalf("zfs recv failed: %v\nOutput:\n%s", err, out)
	}
	exec.Command("udevadm", "settle", "--timeout=5").Run()

	readBack := ddReadBlock(t, "/dev/zvol/"+recvDataset, 16)
	if !bytes.Equal(readBack[:len(testData)], testData) {
		t.Errorf("restored data is not identical (first mismatch at %d)", firstDiff(readBack[:len(testData)], testData))
	}
	assertNoPxitoolSnapshots(t, zvolName)
}

// TestBackupZFSVolume_LiveConcurrentReadersWriters tests a live-backup scenario. A writer continuously rewrites a
// 4K-aligned block of the zvol with a deterministic pattern while multiple readers continuously read the same block.
// The test verifies that the backup completes, the exported image is full-size and self-consistent, and no
// pxitool- snapshot is left behind.
func TestBackupZFSVolume_LiveConcurrentReadersWriters(t *testing.T) {
	const (
		blockSize = 4096
		block     = 16 // 4k-aligned offset: 16 blocks
		numReads  = 4
	)

	poolName := "pxitool2_test_pool_live"
	zvolName := poolName + "/zvol_live"
	zvolTestPool(t, poolName, zvolName, "64M")
	dev := "/dev/zvol/" + zvolName

	// Coherent full-payload image: one 4k block of a deterministic pattern.
	pattern := make([]byte, blockSize)
	for i := range pattern {
		pattern[i] = byte(i % 251)
	}
	writeZvolPattern(t, dev, pattern, block)

	// Verify the pre-fill round-trips before starting the churn.
	if got := ddReadBlock(t, dev, block); !bytes.Equal(got[:blockSize], pattern) {
		t.Fatalf("pre-fill verification failed (first mismatch at %d)", firstDiff(got[:blockSize], pattern))
	}

	stop := make(chan struct{})
	var wg sync.WaitGroup

	// Continuously write the payload block with the same pattern.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			writeZvolPattern(t, dev, pattern, block)
		}
	}()

	// Continuously read the payload block and verify it is the expected pattern.
	for r := 0; r < numReads; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if got := ddReadBlock(t, dev, block); !bytes.Equal(got[:blockSize], pattern) {
					t.Errorf("reader observed a torn volume (first mismatch at %d)", firstDiff(got[:blockSize], pattern))
					return
				}
			}
		}()
	}

	workdir := t.TempDir()
	dest, err := backupZFSVolume(dev, "vol1", workdir)
	close(stop)
	wg.Wait()
	if err != nil {
		t.Fatalf("backupZFSVolume() under live I/O: %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("destination missing: %v", err)
	}

	// Ensure the destination is a full-size coherent copy of the source.
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("backup stream is empty")
	}
	recvDataset := poolName + "/zvol_live_recv"
	t.Cleanup(func() {
		exec.Command("zfs", "destroy", "-R", recvDataset).Run()
	})
	recvCmd := exec.Command("zfs", "recv", "-F", recvDataset)
	recvCmd.Stdin = bytes.NewReader(data)
	if out, err := recvCmd.CombinedOutput(); err != nil {
		t.Fatalf("zfs recv failed: %v\nOutput:\n%s", err, out)
	}
	exec.Command("udevadm", "settle", "--timeout=5").Run()
	readBack := ddReadBlock(t, "/dev/zvol/"+recvDataset, block)
	if !bytes.Equal(readBack[:blockSize], pattern) {
		t.Errorf("restored data is not a coherent snapshot (first mismatch at %d)", firstDiff(readBack[:blockSize], pattern))
	}

	// Ensure no leftover pxitool- snapshot exists on the source dataset.
	assertNoPxitoolSnapshots(t, zvolName)
}

func TestBackupVolume_RoutesZFSPath_Integration(t *testing.T) {
	backingFile, err := os.CreateTemp("", "zfs-test-backing-")
	if err != nil {
		t.Fatal(err)
	}
	if err := backingFile.Truncate(256 * 1024 * 1024); err != nil {
		t.Fatal(err)
	}
	backingFile.Close()
	defer os.Remove(backingFile.Name())

	poolName := "pxitool2_route_pool"
	zvolName := poolName + "/zvol_route"
	runCommandFatal(t, "zpool", "create", "-f", poolName, backingFile.Name())
	t.Cleanup(func() {
		exec.Command("zfs", "destroy", "-R", zvolName).Run()
		exec.Command("zpool", "destroy", "-f", poolName).Run()
	})
	runCommandFatal(t, "zfs", "create", "-V", "64M", zvolName)

	workdir := t.TempDir()
	dest, err := BackupVolume("/dev/zvol/"+zvolName, "vol1", pxi.VolumeTypeZFS, workdir)
	if err != nil {
		t.Fatalf("BackupVolume() error = %v", err)
	}
	want := path.Join(workdir, "vol1")
	if dest != want {
		t.Errorf("dest = %q, want %q", dest, want)
	}
}
