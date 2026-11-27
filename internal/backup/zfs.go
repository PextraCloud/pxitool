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

// backupZFSVolume exports a ZFS zvol to a file in the working directory. The sourcePath must be a valid ZFS zvol path
// (in the form "pool/dataset"). The volumeID is used to name the destination file. Returns the path to the exported file.
func backupZFSVolume(sourcePath, volumeID, workingDirectory string) (string, error) {
	zvol, ok := names.ExtractZFSName(sourcePath)
	if !ok {
		return "", fmt.Errorf("zfs: invalid ZFS zvol path: %s", sourcePath)
	}

	// Generate unique snapshot name
	snapName := fmt.Sprintf("%s@%s", zvol, names.UniqueSnapshotName(volumeID, names.ZFS_MAX_NAME_LEN-len(zvol)-1))
	// Create snapshot
	if out, err := exec.Command("zfs", "snapshot", snapName).CombinedOutput(); err != nil {
		return "", fmt.Errorf("zfs: failed to create snapshot for %s: %w\nOutput:\n%s", zvol, err, string(out))
	}
	defer func() {
		// Destroy snapshot on exit (best-effort)
		if out, err := exec.Command("zfs", "destroy", snapName).CombinedOutput(); err != nil {
			log.Warn("Failed to destroy ZFS snapshot %s: %v\nOutput:\n%s", snapName, err, string(out))
		}
	}()

	// Create destination file in working directory
	destination := path.Join(workingDirectory, volumeID)
	dest, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("zfs: failed to create destination %s: %w", destination, err)
	}
	defer dest.Close()

	// Execute the send
	sendCmd := exec.Command("zfs", "send", snapName)
	// Pipe send output to destination file
	sendCmd.Stdout = dest

	if err := sendCmd.Run(); err != nil {
		// This may cause dest.Close() to fail, but we ignore that error anyways
		os.Remove(destination)
		return "", fmt.Errorf("zfs: failed to send snapshot %s: %w", snapName, err)
	}
	return destination, nil
}
