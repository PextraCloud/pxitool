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
	"fmt"
	"io"
	"os"
	"os/exec"
)

func BackupQEMUVolume(filePath string, writeStream io.Writer) error {
	// A temp file is needed: https://lists.gnu.org/archive/html/qemu-discuss/2020-01/msg00028.html
	// Step 1: Take a copy of the file
	copyFile, err := os.CreateTemp("", "pxitool-qemu-img-copy-*.qcow2")
	if err != nil {
		return fmt.Errorf("failed to create temp copy: %w", err)
	}
	defer os.Remove(copyFile.Name())
	defer copyFile.Close()

	srcFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	if _, err := io.Copy(copyFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy source file: %w", err)
	}

	if err := copyFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync copied file: %w", err)
	}

	// Step 2: Perform convert to remove sparseness
	convertedFile, err := os.CreateTemp("", "pxitool-qemu-img-conv-*.qcow2")
	if err != nil {
		return fmt.Errorf("failed to create temp converted file: %w", err)
	}
	defer os.Remove(convertedFile.Name())
	defer convertedFile.Close()

	cmd := exec.Command("qemu-img", "convert", "-O", "qcow2", "--force-share", copyFile.Name(), convertedFile.Name())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("qemu-img convert failed: %w", err)
	}

	// Step 3: io.Copy the new file and then delete it
	if _, err := convertedFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek converted file: %w", err)
	}
	if _, err := io.Copy(writeStream, convertedFile); err != nil {
		return fmt.Errorf("failed to copy converted qcow2 data: %w", err)
	}

	return nil
}
