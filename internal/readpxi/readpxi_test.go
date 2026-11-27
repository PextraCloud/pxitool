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
package readpxi

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/PextraCloud/pxitool/internal/testutils"
	"github.com/PextraCloud/pxitool/pkg/pxi"
)

const validManifest = `{
  "version": 2,
  "created_by": "pxitool/1.0.0",
  "instance": {
    "id": "inst-1",
    "node_id": "node-1",
    "internal_id": "int-1",
    "type": "qemu",
    "name": "instance-1",
    "metadata": {"_type": "qemu", "data": {}},
    "resources": {
      "cpu": {"sockets": 1, "cores": 2, "threads": 1},
      "memory_bytes": 67108864,
      "volumes": [
        {"id": "vol1", "type": "directory", "storage_pool_id": "pool-1", "path": "/some/path", "size_bytes": 1048576}
      ]
    },
    "autostart": false,
    "boot_order": 0,
    "creation": "2026-01-01T00:00:00Z"
  }
}`

func TestCheckEnvironment(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if err := CheckEnvironment(); err == nil || !strings.Contains(err.Error(), "dar not found in PATH") {
		t.Fatalf("CheckEnvironment() = %v, want error containing %q", err, "dar not found in PATH")
	}

	if err := os.WriteFile(path.Join(dir, "dar"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	if err := CheckEnvironment(); err != nil {
		t.Fatalf("CheckEnvironment() = %v, want nil", err)
	}
}

func TestCreateWorkingDirectory(t *testing.T) {
	dir, err := createWorkingDirectory()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatalf("createWorkingDirectory() dir = %q, stat error = %v", dir, err)
	}
	// Ensure permissions are 0700 (not 0755) to avoid accidental exposure of sensitive data.
	if info.Mode().Perm() != 0o700 {
		t.Errorf("createWorkingDirectory() dir permissions = %o, want 0700", info.Mode().Perm())
	}
}

func TestReadManifest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		wd := t.TempDir()
		if err := os.WriteFile(path.Join(wd, pxi.ManifestFileName), []byte(validManifest), 0o644); err != nil {
			t.Fatal(err)
		}
		m, err := readManifest(wd)
		if err != nil {
			t.Fatal(err)
		}
		if m.Version != 2 || m.Instance.Id != "inst-1" || m.Instance.Name != "instance-1" {
			t.Errorf("unexpected manifest: %+v", m)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if _, err := readManifest(t.TempDir()); err == nil {
			t.Fatal("readManifest() = nil, want error")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		wd := t.TempDir()
		if err := os.WriteFile(path.Join(wd, pxi.ManifestFileName), []byte("not json{"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := readManifest(wd); err == nil {
			t.Fatal("readManifest() = nil, want error")
		}
	})

	t.Run("invalid manifest", func(t *testing.T) {
		wd := t.TempDir()
		bad := strings.Replace(validManifest, `"version": 2`, `"version": 1`, 1)
		if err := os.WriteFile(path.Join(wd, pxi.ManifestFileName), []byte(bad), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := readManifest(wd)
		if err == nil || !strings.Contains(err.Error(), "invalid manifest") {
			t.Fatalf("readManifest() = %v, want error containing %q", err, "invalid manifest")
		}
	})
}

func TestReadPXIHeader(t *testing.T) {
	t.Run("encrypted", func(t *testing.T) {
		testutils.StubScript(t, "dar", "#!/bin/sh\nprintf 'Archive version format               : 11.3\nCompression algorithm used           : zstd\nSymmetric key encryption used        : aes256\n'\n")

		h, err := ReadPXIHeader("fake.pxi")
		if err != nil {
			t.Fatal(err)
		}
		if h.CompressionAlgorithm != "zstd" || h.SymmetricEncryptionAlgorithm != "aes256" || !h.IsEncrypted {
			t.Errorf("unexpected header: %+v", h)
		}
	})

	t.Run("unencrypted", func(t *testing.T) {
		testutils.StubScript(t, "dar", "#!/bin/sh\nprintf 'Archive version format               : 11.3\nCompression algorithm used           : gzip\nSymmetric key encryption used        : none\n'\n")

		h, err := ReadPXIHeader("fake.pxi")
		if err != nil {
			t.Fatal(err)
		}
		if h.CompressionAlgorithm != "gzip" || h.SymmetricEncryptionAlgorithm != "none" || h.IsEncrypted {
			t.Errorf("unexpected header: %+v", h)
		}
	})

	t.Run("unparseable output", func(t *testing.T) {
		testutils.StubScript(t, "dar", "#!/bin/sh\nprintf 'unparseable output'\n")
		if _, err := ReadPXIHeader("fake.pxi"); err == nil {
			t.Fatal("ReadPXIHeader() = nil, want error")
		}
	})
}

func TestReadPXI(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Use dar -R (root directory) as destination to write manifest
		testutils.StubScript(t, "dar", "#!/bin/sh\nfor arg in \"$@\"; do\nif [ \"$arg\" = \"-R\" ]; then\nshift\ndir=\"$1\"\nmkdir -p \"$dir\"\necho '"+validManifest+"' > \"$dir/"+pxi.ManifestFileName+"\"\nexit 0\nfi\nshift\n done\n exit 1\n")
		m, err := ReadPXI("fake.pxi", "")
		if err != nil {
			t.Fatal(err)
		}
		if m.Version != 2 || m.Instance.Id != "inst-1" || m.Instance.Name != "instance-1" {
			t.Errorf("unexpected manifest: %+v", m)
		}
	})

	t.Run("failure", func(t *testing.T) {
		testutils.StubExit(t, "dar", 1)
		if _, err := ReadPXI("fake.pxi", ""); err == nil {
			t.Fatal("ReadPXI() = nil, want error")
		}
	})
}
