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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/PextraCloud/pxitool/internal/testutils"
	"github.com/PextraCloud/pxitool/pkg/pxi"
)

func argsTokens(t *testing.T, logFile string) []string {
	t.Helper()
	raw, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("reading dar args log: %v", err)
	}
	return strings.Fields(string(raw))
}

func hasFlag(tokens []string, flag string) bool {
	for _, tok := range tokens {
		if tok == flag {
			return true
		}
	}
	return false
}

func TestExcludeVolumes(t *testing.T) {
	t.Run("duplicate IDs in exclude list", func(t *testing.T) {
		instance := testInstance(testVolume("vol1", "/tmp/vol1"), testVolume("vol2", "/tmp/vol2"))
		err := excludeVolumes(instance, []string{"vol1", "vol1"})
		if err == nil || !strings.Contains(err.Error(), "duplicate volume ID vol1") {
			t.Fatalf("excludeVolumes() error = %v, want duplicate volume ID error", err)
		}
	})

	t.Run("marks matching volumes excluded", func(t *testing.T) {
		instance := testInstance(testVolume("vol1", "/tmp/vol1"), testVolume("vol2", "/tmp/vol2"))
		if err := excludeVolumes(instance, []string{"vol1"}); err != nil {
			t.Fatalf("excludeVolumes() error = %v", err)
		}
		if !instance.Resources.Volumes[0].Exclude() {
			t.Error("vol1 should be excluded")
		}
		if instance.Resources.Volumes[1].Exclude() {
			t.Error("vol2 should not be excluded")
		}
	})

	t.Run("unknown IDs ignored", func(t *testing.T) {
		instance := testInstance(testVolume("vol1", "/tmp/vol1"))
		if err := excludeVolumes(instance, []string{"nonexistent"}); err != nil {
			t.Fatalf("excludeVolumes() error = %v", err)
		}
		if instance.Resources.Volumes[0].Exclude() {
			t.Error("vol1 should not be excluded")
		}
	})
}

func TestBackupVolumes(t *testing.T) {
	t.Run("all volumes excluded", func(t *testing.T) {
		instance := testInstance(testVolume("vol1", "/tmp/vol1"))
		instance.Resources.Volumes[0].SetExclude(true)
		paths, err := backupVolumes(instance, t.TempDir(), 2)
		if err != nil {
			t.Fatalf("backupVolumes() error = %v", err)
		}
		if len(paths) != 0 {
			t.Errorf("paths = %v, want empty", paths)
		}
	})

	t.Run("no volumes", func(t *testing.T) {
		paths, err := backupVolumes(testInstance(), t.TempDir(), 2)
		if err != nil {
			t.Fatalf("backupVolumes() error = %v", err)
		}
		if len(paths) != 0 {
			t.Errorf("paths = %v, want empty", paths)
		}
	})

	t.Run("success", func(t *testing.T) {
		testutils.StubQemuImg(t)

		srcDir := t.TempDir()

		p1 := filepath.Join(srcDir, "vol1.raw")
		p2 := filepath.Join(srcDir, "vol2.raw")
		if err := os.WriteFile(p1, []byte("volume 1 data"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p2, []byte("volume 2 data"), 0o644); err != nil {
			t.Fatal(err)
		}
		workDir := t.TempDir()
		instance := testInstance(testVolume("vol1", p1), testVolume("vol2", p2))

		paths, err := backupVolumes(instance, workDir, 2)
		if err != nil {
			t.Fatalf("backupVolumes() error = %v", err)
		}
		slices.Sort(paths)
		if !slices.Equal(paths, []string{"vol1", "vol2"}) {
			t.Errorf("paths = %v, want [vol1 vol2]", paths)
		}
		for _, id := range []string{"vol1", "vol2"} {
			if _, err := os.Stat(filepath.Join(workDir, id)); err != nil {
				t.Errorf("backed up %s missing: %v", id, err)
			}
		}
	})

	t.Run("excluded volume skipped", func(t *testing.T) {
		testutils.StubQemuImg(t)

		srcDir := t.TempDir()
		p1 := filepath.Join(srcDir, "vol1.raw")
		if err := os.WriteFile(p1, []byte("volume 1 data"), 0o644); err != nil {
			t.Fatal(err)
		}
		instance := testInstance(testVolume("vol1", p1), testVolume("vol2", "/nonexistent/vol2"))
		instance.Resources.Volumes[1].SetExclude(true)

		paths, err := backupVolumes(instance, t.TempDir(), 2)
		if err != nil {
			t.Fatalf("backupVolumes() error = %v", err)
		}
		if !slices.Equal(paths, []string{"vol1"}) {
			t.Errorf("paths = %v, want [vol1]", paths)
		}
	})

	t.Run("backup failure", func(t *testing.T) {
		instance := testInstance(testVolume("vol1", "/nonexistent/vol1"))
		_, err := backupVolumes(instance, t.TempDir(), 2)
		if err == nil || !strings.Contains(err.Error(), "failed to open source file") {
			t.Fatalf("backupVolumes() error = %v, want failed to open source file", err)
		}
	})
}

func TestCreatePXI(t *testing.T) {
	setup := func(t *testing.T) (logFile, outputPath string) {
		t.Helper()
		tmp := t.TempDir()
		binDir := filepath.Join(tmp, "bin")
		if err := os.Mkdir(binDir, 0755); err != nil {
			t.Fatal(err)
		}
		logFile = filepath.Join(tmp, "dar-args.log")
		testutils.StubQemuImg(t)
		testutils.StubDarLog(t, binDir, logFile)
		outDir := filepath.Join(tmp, "out")
		if err := os.Mkdir(outDir, 0755); err != nil {
			t.Fatal(err)
		}
		return logFile, filepath.Join(outDir, "out.pxi")
	}

	t.Run("success with options", func(t *testing.T) {
		logFile, outputPath := setup(t)
		srcDir := t.TempDir()

		p1 := filepath.Join(srcDir, "vol1.raw")
		p2 := filepath.Join(srcDir, "vol2.raw")
		if err := os.WriteFile(p1, []byte("volume 1 data"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p2, []byte("volume 2 data"), 0o644); err != nil {
			t.Fatal(err)
		}
		instance := testInstance(testVolume("vol1", p1), testVolume("vol2", p2))

		err := CreatePXI(CreatePXIOptions{
			Instance:             instance,
			OutputPath:           outputPath,
			Overwrite:            true,
			EncryptionCipher:     pxi.EncryptionCipherAES256,
			EncryptionKey:        "secret",
			CompressionAlgorithm: pxi.CompressionAlgorithmZstd,
			Concurrency:          0,
		})
		if err != nil {
			t.Fatalf("CreatePXI() error = %v", err)
		}

		tokens := argsTokens(t, logFile)
		for _, want := range []string{"vol1", "vol2", pxi.ManifestFileName, "aes:secret", "zstd", "-w"} {
			if !hasFlag(tokens, want) && !slices.Contains(tokens, want) {
				t.Errorf("dar args %v missing %q", tokens, want)
			}
		}
		if hasFlag(tokens, "-n") {
			t.Errorf("dar args %v should not contain -n with Overwrite=true", tokens)
		}
		if _, err := os.Stat(outputPath); err != nil {
			t.Errorf("output missing: %v", err)
		}
	})

	t.Run("all volumes excluded", func(t *testing.T) {
		logFile, outputPath := setup(t)
		srcDir := t.TempDir()

		p1 := filepath.Join(srcDir, "vol1.raw")
		if err := os.WriteFile(p1, []byte("volume 1 data"), 0o644); err != nil {
			t.Fatal(err)
		}
		instance := testInstance(testVolume("vol1", p1))

		err := CreatePXI(CreatePXIOptions{
			Instance:       instance,
			OutputPath:     outputPath,
			ExcludeVolumes: []string{"vol1"},
			Overwrite:      true,
		})
		if err != nil {
			t.Fatalf("CreatePXI() error = %v", err)
		}
		tokens := argsTokens(t, logFile)
		if !slices.Contains(tokens, pxi.ManifestFileName) {
			t.Errorf("dar args %v missing manifest", tokens)
		}
		if slices.Contains(tokens, "vol1") {
			t.Errorf("dar args %v should not include excluded vol1", tokens)
		}
	})

	t.Run("duplicate exclude list", func(t *testing.T) {
		_, outputPath := setup(t)
		err := CreatePXI(CreatePXIOptions{
			Instance:       testInstance(testVolume("vol1", "/tmp/vol1")),
			OutputPath:     outputPath,
			ExcludeVolumes: []string{"vol1", "vol1"},
		})
		if err == nil || !strings.Contains(err.Error(), "duplicate volume ID vol1") {
			t.Fatalf("CreatePXI() error = %v, want duplicate volume ID error", err)
		}
	})

	t.Run("backup failure", func(t *testing.T) {
		_, outputPath := setup(t)
		instance := testInstance(testVolume("vol1", "/nonexistent/vol1"))
		err := CreatePXI(CreatePXIOptions{
			Instance:   instance,
			OutputPath: outputPath,
		})
		if err == nil || !strings.Contains(err.Error(), "failed to back up volumes") {
			t.Fatalf("CreatePXI() error = %v, want failed to back up volumes", err)
		}
	})

	t.Run("overwrite false", func(t *testing.T) {
		logFile, outputPath := setup(t)
		srcDir := t.TempDir()

		p1 := filepath.Join(srcDir, "vol1.raw")
		if err := os.WriteFile(p1, []byte("volume 1 data"), 0o644); err != nil {
			t.Fatal(err)
		}
		instance := testInstance(testVolume("vol1", p1))

		err := CreatePXI(CreatePXIOptions{
			Instance:   instance,
			OutputPath: outputPath,
			Overwrite:  false,
		})
		if err != nil {
			t.Fatalf("CreatePXI() error = %v", err)
		}
		tokens := argsTokens(t, logFile)
		if !hasFlag(tokens, "-n") {
			t.Errorf("dar args %v missing -n", tokens)
		}
	})

	t.Run("negative concurrency clamped", func(t *testing.T) {
		_, outputPath := setup(t)
		srcDir := t.TempDir()

		p1 := filepath.Join(srcDir, "vol1.raw")
		if err := os.WriteFile(p1, []byte("volume 1 data"), 0o644); err != nil {
			t.Fatal(err)
		}
		instance := testInstance(testVolume("vol1", p1))

		err := CreatePXI(CreatePXIOptions{
			Instance:    instance,
			OutputPath:  outputPath,
			Concurrency: -5,
			Overwrite:   true,
		})
		if err != nil {
			t.Fatalf("CreatePXI() error = %v", err)
		}
	})

	t.Run("manifest invalid", func(t *testing.T) {
		_, outputPath := setup(t)
		instance := testInstance()
		instance.Id = ""
		err := CreatePXI(CreatePXIOptions{
			Instance:   instance,
			OutputPath: outputPath,
		})
		if err == nil || !strings.Contains(err.Error(), "failed to prepare PXI") {
			t.Fatalf("CreatePXI() error = %v, want failed to prepare PXI", err)
		}
	})
}

func TestCheckEnvironment(t *testing.T) {
	t.Run("both tools present", func(t *testing.T) {
		testutils.StubExit(t, "dar", 0)
		testutils.StubQemuImg(t)
		if err := CheckEnvironment(); err != nil {
			t.Fatalf("CheckEnvironment() error = %v", err)
		}
	})

	t.Run("missing tools", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		err := CheckEnvironment()
		if err == nil {
			t.Fatal("CheckEnvironment() = nil, want error")
		}
		if !strings.Contains(err.Error(), "dar not found in PATH") || !strings.Contains(err.Error(), "qemu-img not found in PATH") {
			t.Errorf("CheckEnvironment() error = %q, want both tool names", err)
		}
	})
}
