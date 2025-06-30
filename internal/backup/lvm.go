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
	"strings"
	"time"

	"github.com/PextraCloud/pxitool/pkg/log"
)

func pathToLVMVolume(volumePath string) (string, string) { // vgname, lvname
	// format: /dev/[vgname]/[lvname]
	if len(volumePath) < 5 || volumePath[:5] != "/dev/" {
		return "", ""
	}
	// Remove /dev/ prefix
	volumePath = volumePath[5:]
	// Split by slash
	parts := strings.Split(volumePath, "/")
	if len(parts) != 2 {
		return "", ""
	}
	vgName := parts[0]
	lvName := parts[1]

	return vgName, lvName
}

func BackupLVMVolume(volumePath string, writeStream io.Writer) error {
	vgName, lvName := pathToLVMVolume(volumePath)

	// Create snapshot
	timestamp := time.Now().Format("20060102150405")
	snapshotName := fmt.Sprintf("%s-pxitool-%s", lvName, timestamp)
	createCmd := exec.Command("lvcreate", "--snapshot", "--name", snapshotName, "--size", "256M", fmt.Sprintf("/dev/%s/%s", vgName, lvName))

	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create LVM snapshot: %w", err)
	}
	// Destroy snapshot after sending
	defer func() {
		deleteCmd := exec.Command("lvremove", "-f", fmt.Sprintf("/dev/%s/%s", vgName, snapshotName))
		if err := deleteCmd.Run(); err != nil {
			log.Warn("Failed to destroy LVM snapshot %s: %v", snapshotName, err)
		}
	}()

	cmd := exec.Command("dd", fmt.Sprintf("if=/dev/%s/%s", vgName, snapshotName), "bs=8M", "status=none")
	cmd.Stderr = io.MultiWriter(writeStream, io.Discard, io.Writer(os.Stderr)) // Do not specify "of=" so it writes to stdout
	cmd.Stdout = writeStream
	return cmd.Run()
}
