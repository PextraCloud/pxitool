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
	"path/filepath"
	"strings"
	"testing"

	"github.com/PextraCloud/pxitool/internal/testutils"
	"github.com/PextraCloud/pxitool/internal/version"
	"github.com/PextraCloud/pxitool/pkg/pxi"
)

func TestGetManifest(t *testing.T) {
	t.Run("dar unavailable", func(t *testing.T) {
		t.Setenv("PATH", "")
		instance := testInstance(testVolume("vol1", "/tmp/vol1"))
		manifest := getManifest(instance)
		if manifest.Version != pxi.ManifestVersion2 {
			t.Errorf("version = %d, want %d", manifest.Version, pxi.ManifestVersion2)
		}
		want := "pxitool " + version.PxiToolVersion + " (unknown (error retrieving dar version))"
		if manifest.CreatedBy != want {
			t.Errorf("createdBy = %q, want %q", manifest.CreatedBy, want)
		}
		if manifest.Instance.Id != instance.Id {
			t.Errorf("instance id = %q, want %q", manifest.Instance.Id, instance.Id)
		}
	})

	t.Run("dar available", func(t *testing.T) {
		testutils.StubScript(t, "dar", "#!/bin/sh\nprintf '  dar 3.0.0\\nlibdar 2.7.2\\n'\n")

		instance := testInstance()
		manifest := getManifest(instance)
		want := "pxitool " + version.PxiToolVersion + " (dar 3.0.0, libdar 2.7.2)"
		if manifest.CreatedBy != want {
			t.Errorf("createdBy = %q, want %q", manifest.CreatedBy, want)
		}
	})
}

func TestWriteManifest(t *testing.T) {
	t.Run("valid manifest written", func(t *testing.T) {
		t.Setenv("PATH", "")
		instance := testInstance(testVolume("vol1", "/tmp/vol1"))
		manifest := getManifest(instance)

		path := filepath.Join(t.TempDir(), pxi.ManifestFileName)
		if err := writeManifest(manifest, path); err != nil {
			t.Fatalf("writeManifest() error = %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var decoded pxi.Manifest
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("manifest is not valid JSON: %v", err)
		}
		if decoded.Version != pxi.ManifestVersion2 {
			t.Errorf("version = %d, want %d", decoded.Version, pxi.ManifestVersion2)
		}
		if decoded.Instance.Id != instance.Id {
			t.Errorf("instance id = %q, want %q", decoded.Instance.Id, instance.Id)
		}
		if len(decoded.Instance.Resources.Volumes) != 1 || decoded.Instance.Resources.Volumes[0].Id != "vol1" {
			t.Errorf("volumes = %+v, want vol1", decoded.Instance.Resources.Volumes)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o644 {
			t.Errorf("perm = %o, want 644", info.Mode().Perm())
		}
	})

	t.Run("invalid manifest", func(t *testing.T) {
		t.Setenv("PATH", "")
		instance := testInstance()
		instance.Id = "" // invalid
		manifest := getManifest(instance)

		path := filepath.Join(t.TempDir(), pxi.ManifestFileName)
		err := writeManifest(manifest, path)
		if err == nil || !strings.Contains(err.Error(), "invalid manifest") {
			t.Fatalf("writeManifest() error = %v, want invalid manifest error", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("manifest file should not exist, stat err = %v", err)
		}
	})

	t.Run("unsupported version", func(t *testing.T) {
		t.Setenv("PATH", "")
		instance := testInstance()
		manifest := getManifest(instance)
		manifest.Version = 1

		path := filepath.Join(t.TempDir(), pxi.ManifestFileName)
		err := writeManifest(manifest, path)
		if err == nil || !strings.Contains(err.Error(), "invalid manifest") {
			t.Fatalf("writeManifest() error = %v, want invalid manifest error", err)
		}
	})

	t.Run("unwritable path", func(t *testing.T) {
		t.Setenv("PATH", "")
		instance := testInstance()
		manifest := getManifest(instance)
		path := filepath.Join(t.TempDir(), "nonexistent-dir", pxi.ManifestFileName)
		err := writeManifest(manifest, path)
		if err == nil || !strings.Contains(err.Error(), "failed to write manifest") {
			t.Fatalf("writeManifest() error = %v, want failed to write manifest error", err)
		}
	})
}
