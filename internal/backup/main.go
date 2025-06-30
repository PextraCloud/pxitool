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

	"github.com/PextraCloud/pxitool/internal/utils"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/volumetype"
)

type VolumeBackupPayload struct {
	Path string
	Type volumetype.VolumeType
}

// Backs up a volume based on its type
func BackupVolume(volumePath string, volumeType volumetype.VolumeType, writeStream io.Writer) (int64, error) {
	countingWriter := utils.NewCountingWriter(writeStream)
	switch volumeType {
	case volumetype.Directory, volumetype.NetFS:
		err := BackupQEMUVolume(volumePath, countingWriter)
		return countingWriter.Count(), err
	case volumetype.LVM:
		err := BackupLVMVolume(volumePath, countingWriter)
		return countingWriter.Count(), err
	case volumetype.ZFS:
		err := BackupZFSVolume(volumePath, countingWriter)
		return countingWriter.Count(), err
	case volumetype.RBD:
		err := BackupRBDVolume(volumePath, countingWriter)
		return countingWriter.Count(), err
	case volumetype.ISCSI:
		return 0, fmt.Errorf("iSCSI volumes should not be backed up directly; use the underlying block device")
	default:
		return 0, fmt.Errorf("unsupported volume type: %s", volumeType)
	}
}
