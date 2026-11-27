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
	"path"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/PextraCloud/pxitool/internal/testutils"
)

// lxcTestRootfs builds a throwaway rootfs directory tree with a known contents marker file and registers cleanup.
// Returns the rootfs path.
func lxcTestRootfs(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, rel := range []string{"etc", "var/lib", "sub"} {
		if err := os.MkdirAll(path.Join(dir, rel), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
	}
	if err := os.WriteFile(path.Join(dir, "marker.txt"), []byte("pxitool2 lxc rootfs test data"), 0644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	return dir
}

func TestBackupLXCRootfsVolume_SourceMissing(t *testing.T) {
	_, err := backupLXCRootfsVolume(path.Join(t.TempDir(), "nope"), "vol1", t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	if !strings.Contains(err.Error(), "failed to stat source") {
		t.Errorf("error = %v, want 'failed to stat source'", err)
	}
}

func TestBackupLXCRootfsVolume_SourceNotADir(t *testing.T) {
	file := path.Join(t.TempDir(), "afile")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := backupLXCRootfsVolume(file, "vol1", t.TempDir())
	if err == nil {
		t.Fatal("expected error for non-directory source")
	}
	if !strings.Contains(err.Error(), "is not a directory") {
		t.Errorf("error = %v, want 'is not a directory'", err)
	}
}

func TestBackupLXCRootfsVolume_TarExitsOne_Success(t *testing.T) {
	src := lxcTestRootfs(t)
	workdir := t.TempDir()
	testutils.StubExit(t, "tar", 1) // simulate tar exit 1 (non-fatal)

	dest, err := backupLXCRootfsVolume(src, "vol1", workdir)
	if err != nil {
		t.Fatalf("backupLXCRootfsVolume() error = %v, want success for non-fatal tar exit 1", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("destination missing: %v", err)
	}
}

func TestBackupLXCRootfsVolume_TarFails_RemovesPartialDest(t *testing.T) {
	src := lxcTestRootfs(t)
	workdir := t.TempDir()
	testutils.StubExit(t, "tar", 2) // simulate tar exit 2 (fatal)

	_, err := backupLXCRootfsVolume(src, "vol1", workdir)
	if err == nil {
		t.Fatal("expected error on tar failure")
	}
	if !strings.Contains(err.Error(), "tar failed") {
		t.Errorf("error = %v, want 'tar failed'", err)
	}

	entries, _ := os.ReadDir(workdir)
	if got := dirNames(entries); len(got) != 0 {
		t.Errorf("working directory after failure = %v, want empty", got)
	}
}

func TestBackupLXCRootfsVolume_Success(t *testing.T) {
	src := lxcTestRootfs(t)
	workdir := t.TempDir()

	dest, err := backupLXCRootfsVolume(src, "vol1", workdir)
	if err != nil {
		t.Fatalf("backupLXCRootfsVolume() error = %v", err)
	}
	if want := path.Join(workdir, "vol1"); dest != want {
		t.Errorf("dest = %q, want %q", dest, want)
	}

	names := tarList(t, dest)
	for _, want := range []string{"./marker.txt", "./etc", "./var/lib", "./sub"} {
		if !slices.Contains(names, want) {
			t.Errorf("archive missing %s; got %v", want, names)
		}
	}
	got := tarExtract(t, dest, "./marker.txt")
	if !bytes.Equal(got, []byte("pxitool2 lxc rootfs test data")) {
		t.Errorf("marker contents = %q, want the test marker", got)
	}
}

func TestBackupLXCRootfsVolume_LiveConcurrentWriters(t *testing.T) {
	src := lxcTestRootfs(t)
	marker := path.Join(src, "marker.txt")

	// Deterministic pattern; the writer rewrites whole 4K-aligned blocks with this identical pattern.
	const size = 4096
	pattern := make([]byte, size)
	for i := range pattern {
		pattern[i] = byte(i % 251)
	}
	if err := os.WriteFile(marker, pattern, 0644); err != nil {
		t.Fatal(err)
	}

	stop := make(chan struct{})
	var wg sync.WaitGroup

	// Writers: continuously rewrite the payload block with the same pattern.
	wg.Add(1)
	go func() {
		defer wg.Done()
		f, err := os.OpenFile(marker, os.O_WRONLY, 0)
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
			if _, err := f.WriteAt(pattern, 0); err != nil {
				t.Errorf("writer: write: %v", err)
				return
			}
		}
	}()

	dest, err := backupLXCRootfsVolume(src, "vol1", t.TempDir())
	close(stop)
	wg.Wait()
	if err != nil {
		t.Fatalf("backupLXCRootfsVolume() under live I/O: %v", err)
	}

	names := tarList(t, dest)
	if !slices.Contains(names, "./marker.txt") {
		t.Fatalf("archive missing ./marker.txt; got %v", names)
	}

	// Verify the pattern
	got := tarExtract(t, dest, "./marker.txt")
	if len(got) != size {
		t.Errorf("marker len = %d, want %d (torn/truncated write)", len(got), size)
	}
	if i := firstDiff(got, pattern); i >= 0 {
		t.Errorf("marker is not a coherent pattern (first mismatch at %d)", i)
	}
}
