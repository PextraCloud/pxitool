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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/PextraCloud/pxitool/internal/volumeformat"
)

// isFilesystemPath reports whether p is a valid filesystem path.
// Also used by lxc_rootfs.go.
func isFilesystemPath(p string) bool {
	if p == "." || p == "/" {
		// Empty or root paths not allowed
		return false
	} else if !path.IsAbs(p) {
		// Relative paths not allowed
		return false
	} else if len(p) > syscall.PathMax {
		// Path too long
		return false
	}
	return true
}

// backupQEMUVolume creates a backup of a QEMU volume.
func backupQEMUVolume(sourcePath, volumeID, workingDirectory string) (string, error) {
	// Clean and validate sourcePath
	sourcePath = path.Clean(sourcePath)
	if !isFilesystemPath(sourcePath) {
		return "", fmt.Errorf("qemu: source path is not a valid filesystem path: %s", sourcePath)
	}

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("qemu: failed to open source file: %w", err)
	}
	defer srcFile.Close()

	var magic [4]byte
	if _, err := io.ReadFull(srcFile, magic[:]); err != nil {
		return "", fmt.Errorf("qemu: failed to read source file header: %w", err)
	}
	format := volumeformat.GetVolumeFormat(magic)

	// Create destination file in working directory
	destination := path.Join(workingDirectory, volumeID)
	output, err := exec.Command("qemu-img", "convert",
		"-f", format.String(),
		"-O", format.String(),
		"--force-share",
		sourcePath, destination).CombinedOutput()

	// Clean up on error
	if err != nil {
		os.Remove(destination)
		return "", fmt.Errorf("qemu: qemu-img convert failed: %w\nOutput:\n%s", err, output)
	}
	return destination, nil
}
