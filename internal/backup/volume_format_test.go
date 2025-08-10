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
	"os"
	"os/exec"
	"testing"

	"github.com/PextraCloud/pxitool/pkg/pxi/constants/volumeformat"
)

// createAndReadSignature creates a temporary image file using qemu-img,
// reads its first 4 bytes, and then cleans up the file.
func createAndReadSignature(t *testing.T, format string) ([4]byte, bool) {
	t.Helper()
	if !commandExists("qemu-img") {
		t.Skipf("Skipping test: command 'qemu-img' not found")
		return [4]byte{}, false
	}

	tmpfile, err := os.CreateTemp("", "pxitool-test-sig-*.img")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	fileName := tmpfile.Name()
	tmpfile.Close()
	defer os.Remove(fileName)

	cmd := exec.Command("qemu-img", "create", "-f", format, fileName, "1M")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create %s image: %v", format, err)
	}

	file, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("Failed to open created image: %v", err)
	}
	defer file.Close()

	var signature [4]byte
	if _, err := file.Read(signature[:]); err != nil {
		t.Fatalf("Failed to read signature from image: %v", err)
	}
	return signature, true
}

func TestGetVolumeFormat(t *testing.T) {
	t.Run("QCOW2 signature", func(t *testing.T) {
		signature, ok := createAndReadSignature(t, "qcow2")
		if !ok {
			return
		}
		result := getVolumeFormat(signature)
		if result != volumeformat.QCOW2 {
			t.Errorf("expected volume format %q, but got %q", volumeformat.QCOW2, result)
		}
	})
	t.Run("VMDK signature", func(t *testing.T) {
		signature, ok := createAndReadSignature(t, "vmdk")
		if !ok {
			return
		}
		result := getVolumeFormat(signature)
		if result != volumeformat.VMDK {
			t.Errorf("expected volume format %q, but got %q", volumeformat.VMDK, result)
		}
	})
	t.Run("Unknown signature (zeros)", func(t *testing.T) {
		input := [4]byte{0x00, 0x00, 0x00, 0x00}
		expected := volumeformat.Raw
		result := getVolumeFormat(input)
		if result != expected {
			t.Errorf("expected volume format %q, but got %q", expected, result)
		}
	})
	t.Run("Unknown signature (random)", func(t *testing.T) {
		input := [4]byte{0x12, 0x34, 0x56, 0x78}
		expected := volumeformat.Raw
		result := getVolumeFormat(input)
		if result != expected {
			t.Errorf("expected volume format %q, but got %q", expected, result)
		}
	})
}
