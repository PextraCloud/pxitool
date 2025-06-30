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
	"os/exec"
	"time"
)

func pathToZFSDataset(volumePath string) string {
	// format: /dev/zvol/poolname/datasetname
	if len(volumePath) < 10 || volumePath[:10] != "/dev/zvol/" {
		return volumePath
	}
	// Remove /dev/zvol/ prefix
	datasetName := volumePath[10:]
	return datasetName
}

func BackupZFSVolume(volumePath string, writeStream io.Writer) error {
	dataset := pathToZFSDataset(volumePath)

	timestamp := time.Now().Format("20060102150405")
	snapshotName := fmt.Sprintf("%s@pxitool_%s", dataset, timestamp)
	fmt.Printf("Creating ZFS snapshot: %s\n", snapshotName)

	// Create snapshot
	createCmd := exec.Command("zfs", "snapshot", snapshotName)
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}
	// Destroy snapshot after sending
	defer func() {
		deleteCmd := exec.Command("zfs", "destroy", snapshotName)
		_ = deleteCmd.Run()
	}()

	cmd := exec.Command("zfs", "send", snapshotName)
	cmd.Stdout = writeStream
	return cmd.Run()
}
