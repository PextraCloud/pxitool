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
	"os"
	"os/exec"
	"path"

	"github.com/PextraCloud/pxitool/internal/names"
	"github.com/PextraCloud/pxitool/pkg/log"
)

// backupRBDVolume exports an RBD image to a file in the working directory. The sourcePath must be a valid RBD image path
// (in the form "pool/image"). The volumeID is used to name the destination file. Returns the path to the exported file.
func backupRBDVolume(sourcePath, volumeID, workingDirectory string) (string, error) {
	pool, image, ok := names.ExtractRBDNames(sourcePath)
	if !ok {
		return "", fmt.Errorf("rbd: unsupported RBD image path: %q", sourcePath)
	}

	// Generate unique snapshot name
	snapName := names.UniqueSnapshotName(volumeID, names.RBD_MAX_NAME_LEN)
	// Create snapshot
	if out, err := exec.Command("rbd", "snap", "create", "--snap", snapName, image).CombinedOutput(); err != nil {
		return "", fmt.Errorf("rbd: failed to create snapshot for %s: %w\nOutput:\n%s", image, err, string(out))
	}
	defer func() {
		// Destroy snapshot on exit (best-effort); unprotect first, then remove.
		exec.Command("rbd", "snap", "unprotect", "--snap", snapName, image).Run()
		if out, err := exec.Command("rbd", "snap", "rm", "--snap", snapName, image).CombinedOutput(); err != nil {
			log.Warn("Failed to remove RBD snapshot %s/%s@%s: %v\nOutput:\n%s", pool, image, snapName, err, string(out))
		}
	}()

	// Create destination file in working directory
	destination := path.Join(workingDirectory, volumeID)
	dest, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("rbd: failed to create destination %s: %w", destination, err)
	}
	defer dest.Close()

	// Execute the export
	exportCmd := exec.Command("rbd", "export", image+"@"+snapName)
	// Pipe export output to destination file
	exportCmd.Stdout = dest

	if err := exportCmd.Run(); err != nil {
		// This may cause dest.Close() to fail, but we ignore that error anyways
		os.Remove(destination)
		return "", fmt.Errorf("rbd: failed to export snapshot %s@%s: %w", image, snapName, err)
	}
	return destination, nil
}
