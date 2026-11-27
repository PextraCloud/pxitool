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

	"github.com/PextraCloud/pxitool/internal/execu"
)

// backupLXCRootfsVolume backs up an LXC container rootfs (a directory) by streaming it with `tar` into the destination
// file.
func backupLXCRootfsVolume(sourcePath, volumeID, workingDirectory string) (string, error) {
	// Clean and validate sourcePath
	sourcePath = path.Clean(sourcePath)
	if !isFilesystemPath(sourcePath) {
		return "", fmt.Errorf("lxc rootfs: source path is not a valid filesystem path: %s", sourcePath)
	} else if fi, err := os.Stat(sourcePath); err != nil {
		return "", fmt.Errorf("lxc rootfs: failed to stat source %s: %w", sourcePath, err)
	} else if !fi.IsDir() {
		// Avoid archiving a non-directory (e.g. file, symlink, block device)
		return "", fmt.Errorf("lxc rootfs: source %s is not a directory", sourcePath)
	}

	// Create destination file in working directory
	destination := path.Join(workingDirectory, volumeID)
	dest, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("lxc rootfs: failed to create destination %s: %w", destination, err)
	}
	defer dest.Close()

	cmd := execu.Command("tar",
		// Stay on the rootfs, do not archive /proc, /sys, /dev, etc.
		"--one-file-system",
		// Preserve permissions, sparse files, numeric owner, and ACLs
		"-p", "--sparse", "--numeric-owner", "--acls",
		// Preserve user.* and security.capability xattrs (file capabilities)
		"--xattrs", "--xattrs-include=user.*", "--xattrs-include=security.capability",
		// Ignore "file changed as we read it", "file ignored", and "xattr write failed" warnings on a live rootfs
		"--warning=no-file-changed", "--warning=no-file-ignored", "--warning=no-xattr-write",
		"-C", sourcePath,
		"-cf", "-", ".")
	// Write tar output to destination file
	cmd.Stdout = dest
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	// Clean up destination file on error
	if err != nil {
		var exitErr *exec.ExitError
		var ok bool
		if exitErr, ok = err.(*exec.ExitError); !ok {
			// Unexpected error, return
			// This may cause dest.Close() to fail, but we ignore that error anyways
			os.Remove(destination)
			return "", fmt.Errorf("lxc rootfs: tar failed: %w", err)
		}

		if exitErr.ExitCode() == 0 || exitErr.ExitCode() == 1 {
			// tar exited with 0 or 1, which is success or non-fatal warning (see tar(1)).
			return destination, nil
		}

		// tar exited with 2 (fatal error), clean
		// This may cause dest.Close() to fail, but we ignore that error anyways
		os.Remove(destination)
		return "", fmt.Errorf("lxc rootfs: tar failed: %w", err)
	}
	return destination, nil
}
