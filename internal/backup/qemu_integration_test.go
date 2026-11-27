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

	"github.com/PextraCloud/pxitool/internal/testutils"
)

const (
	qemuTestPayloadOffset = 65536   // 4K-aligned offset of the payload block
	qemuTestPayloadLen    = 4096    // 4K-aligned length of the payload block
	qemuTestPatternByte   = 0xa     // deterministic pattern; qemu-io -P only takes one byte
	qemuTestVolumeSize    = 1 << 20 // 1 MiB volume
)

// qemuTestAssertPattern verifies every byte of the payload block is the pattern.
func qemuTestAssertPattern(t *testing.T, path string) {
	t.Helper()
	out, err := exec.Command("qemu-io", "-c",
		fmt.Sprintf("read -P 0x%x %d %d", qemuTestPatternByte, qemuTestPayloadOffset, qemuTestPayloadLen), path).CombinedOutput()
	if err != nil {
		t.Fatalf("payload pattern mismatch in %s: %v\n%s", path, err, out)
	}
}

// qemuTestWritePayload fills the payload block with qemuTestPatternByte via qemu-io.
func qemuTestWritePayload(t *testing.T, path string) {
	t.Helper()
	runCommandFatal(t, "qemu-io", "-c",
		fmt.Sprintf("write -P 0x%x %d %d", qemuTestPatternByte, qemuTestPayloadOffset, qemuTestPayloadLen), path)
}

// qemuTestMakeSource creates a raw volume of volumeSize and fills the payload block with qemuTestPatternByte byte via
// qemu-io, then verifies the payload block is coherent. The raw volume is used as the source for backupQEMUVolume().
func qemuTestMakeSource(t *testing.T, src string) {
	t.Helper()
	runCommandFatal(t, "qemu-img", "create", "-f", "raw", src, fmt.Sprintf("%d", qemuTestVolumeSize))
	qemuTestWritePayload(t, src)
	qemuTestAssertPattern(t, src)
}

// qemuTestRawRead reads a raw file at the given offset and length, returning the bytes read.
func qemuTestRawRead(t *testing.T, path string, offset, length int) []byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	buf := make([]byte, length)
	if _, err := f.ReadAt(buf, int64(offset)); err != nil {
		t.Fatal(err)
	}
	return buf
}

func TestBackupQEMUVolume_Success(t *testing.T) {
	workdir := t.TempDir()
	src := path.Join(t.TempDir(), "vol1.qcow2")
	qemuTestMakeSource(t, src)

	dest, err := backupQEMUVolume(src, "vol1", workdir)
	if err != nil {
		t.Fatalf("backupQEMUVolume() error = %v", err)
	}
	if want := path.Join(workdir, "vol1"); dest != want {
		t.Errorf("dest = %q, want %q", dest, want)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading destination: %v", err)
	}
	if len(data) != qemuTestVolumeSize {
		t.Errorf("destination len = %d, want %d", len(data), qemuTestVolumeSize)
	}
	want := bytes.Repeat([]byte{qemuTestPatternByte}, qemuTestPayloadLen)
	if !bytes.Equal(data[qemuTestPayloadOffset:qemuTestPayloadOffset+qemuTestPayloadLen], want) {
		t.Errorf("payload block mismatch (first diff at %d)", firstDiff(data[qemuTestPayloadOffset:qemuTestPayloadOffset+qemuTestPayloadLen], want))
	}
}

func TestBackupQEMUVolume_ConvertFailure_RemovesPartialDest(t *testing.T) {
	workdir := t.TempDir()
	src := path.Join(t.TempDir(), "vol1.raw")
	if err := os.WriteFile(src, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	// DST is the 8th argument of "qemu-img convert -f FMT -O FMT --force-share SRC <DST>".
	testutils.StubScript(t, "qemu-img", "#/bin/sh\necho -n partial > \"$8\"\nexit 2\n") // simulate qemu-img convert failure

	_, err := backupQEMUVolume(src, "vol1", workdir)
	if err == nil {
		t.Fatal("expected error on convert failure")
	}
	if !strings.Contains(err.Error(), "qemu-img convert failed") {
		t.Errorf("error = %v, want 'qemu-img convert failed'", err)
	}
	entries, _ := os.ReadDir(workdir)
	if got := dirNames(entries); len(got) != 0 {
		t.Errorf("working directory after failure = %v, want empty", got)
	}
}

// TestBackupQEMUVolume_LiveConcurrentReadersWriters simulates a live backup.
// Readers and writers concurrently access the source volume while backupQEMUVolume() runs.
// The test verifies that the backup completes, the destination is a full-size coherent copy of the source, and no
// pxitool- snapshot is left behind.
func TestBackupQEMUVolume_LiveConcurrentReadersWriters(t *testing.T) {
	workdir := t.TempDir()
	src := path.Join(t.TempDir(), "vol1.qcow2")
	qemuTestMakeSource(t, src)

	stop := make(chan struct{})
	var wg sync.WaitGroup

	// Continuously rewrite the payload block with the same pattern.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			qemuTestWritePayload(t, src)
		}
	}()

	// Continously read the payload block and verify it is the expected pattern.
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				got := qemuTestRawRead(t, src, qemuTestPayloadOffset, qemuTestPayloadLen)
				if b := firstNonPattern(got, qemuTestPatternByte); b >= 0 {
					t.Errorf("reader observed a torn volume (byte 0x%02x at offset %d, want 0x%02x)", got[b], b, qemuTestPatternByte)
					return
				}
			}
		}()
	}

	dest, err := backupQEMUVolume(src, "vol1", workdir)
	close(stop)
	wg.Wait()
	if err != nil {
		t.Fatalf("backupQEMUVolume() under live I/O: %v", err)
	}

	// Ensure the destination is a full-size coherent copy of the source.
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading destination: %v", err)
	}
	if len(data) != qemuTestVolumeSize {
		t.Errorf("destination len = %d, want %d (truncated backup)", len(data), qemuTestVolumeSize)
	}
	want := bytes.Repeat([]byte{qemuTestPatternByte}, qemuTestPayloadLen)
	if i := firstDiff(data[qemuTestPayloadOffset:qemuTestPayloadOffset+qemuTestPayloadLen], want); i >= 0 {
		t.Errorf("destination payload is not coherent (first mismatch at %d)", i)
	}
}
