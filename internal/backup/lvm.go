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

// lvmSnapshotSize is the size of the COW snapshot area used for backups. It must be large enough to absorb writes to
// the origin during the backup, otherwise, the snapshot will run out of space and the backup will fail.
// 64M is a reasonable default for most workloads.
const lvmSnapshotSize = "64M"

// backupLVMVolume backs up an LVM logical volume (sourcePath of the form /dev/<vg>/<lv> or /dev/mapper/<vg>-<lv>) by
// creating a short-lived snapshot and reading it with `dd`.
//
// NOTE: During the backup, writes to the origin volume must not exceed the size of the snapshot
// (lvmSnapshotSize). If the snapshot runs out of space, the backup will fail. This is a current limitation of
// LVM snapshots.
func backupLVMVolume(sourcePath, volumeID, workingDirectory string) (string, error) {
	vg, lv, ok := names.ExtractLVMNames(sourcePath)
	if !ok {
		return "", fmt.Errorf("lvm: unsupported LVM volume path: %q", sourcePath)
	}

	// Generate unique snapshot name
	snapName := names.UniqueSnapshotName(volumeID, names.LVM_MAX_NAME_LEN)
	// Create snapshot
	if out, err := exec.Command("lvcreate", "--snapshot", "--name", snapName,
		"--size", lvmSnapshotSize, fmt.Sprintf("/dev/%s/%s", vg, lv)).CombinedOutput(); err != nil {
		return "", fmt.Errorf("lvm: failed to create snapshot for %s: %w\nOutput:\n%s", sourcePath, err, string(out))
	}
	defer func() {
		if out, err := exec.Command("lvremove", "-f", fmt.Sprintf("/dev/%s/%s", vg, snapName)).CombinedOutput(); err != nil {
			log.Warn("Failed to destroy LVM snapshot %s/%s: %v\nOutput:\n%s", vg, snapName, err, string(out))
		}
	}()

	// Create destination file in working directory
	destination := path.Join(workingDirectory, volumeID)
	dest, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("lvm: failed to create destination %s: %w", destination, err)
	}
	defer dest.Close()

	// Execute the read/write
	ddCmd := exec.Command("dd", fmt.Sprintf("if=/dev/%s/%s", vg, snapName), "bs=8M", "status=none")
	// Pipe output to destination file
	ddCmd.Stdout = dest

	if err := ddCmd.Run(); err != nil {
		// This may cause dest.Close() to fail, but we ignore that error anyways
		os.Remove(destination)
		return "", fmt.Errorf("lvm: failed to read snapshot %s/%s: %w", vg, snapName, err)
	}
	return destination, nil
}
