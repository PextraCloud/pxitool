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
package createpxi

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PextraCloud/pxitool/pkg/pxi"
)

func TestCreateWorkingDirectory(t *testing.T) {
	dir, err := createWorkingDirectory()
	if err != nil {
		t.Fatalf("createWorkingDirectory() error = %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("working directory missing: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("%s is not a directory", dir)
	}
	if !strings.HasPrefix(path.Base(dir), "pxi-create-") {
		t.Errorf("dir base = %q, want prefix pxi-create-", path.Base(dir))
	}
}

func TestPreparePXI(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Setenv("PATH", "")
		instance := testInstance(testVolume("vol1", "/tmp/vol1"))

		dir, err := preparePXI(instance)
		if err != nil {
			t.Fatalf("preparePXI() error = %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(dir) })

		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("working directory missing: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(dir, pxi.ManifestFileName))
		if err != nil {
			t.Fatalf("reading manifest: %v", err)
		}
		var manifest pxi.Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			t.Fatalf("manifest is not valid JSON: %v", err)
		}
		if manifest.Version != pxi.ManifestVersion2 {
			t.Errorf("version = %d, want %d", manifest.Version, pxi.ManifestVersion2)
		}
		if manifest.Instance.Id != instance.Id {
			t.Errorf("instance id = %q, want %q", manifest.Instance.Id, instance.Id)
		}
	})

	t.Run("invalid instance", func(t *testing.T) {
		t.Setenv("PATH", "")
		instance := testInstance()
		instance.Id = ""

		dir, err := preparePXI(instance)
		if err == nil {
			t.Fatal("preparePXI() = nil, want error")
		}
		t.Cleanup(func() {
			if dir != "" {
				os.RemoveAll(dir)
			}
		})
	})
}
