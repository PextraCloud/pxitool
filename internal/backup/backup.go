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

	"github.com/PextraCloud/pxitool/pkg/pxi"
)

// BackupVolume backs up a volume to the specified working directory, returning the path to the backup file.
func BackupVolume(sourcePath string, volumeID string, volumeType pxi.VolumeType, workingDirectory string) (string, error) {
	switch volumeType {
	case pxi.VolumeTypeDirectory:
		return backupQEMUVolume(sourcePath, volumeID, workingDirectory)
	case pxi.VolumeTypeLVM:
		return backupLVMVolume(sourcePath, volumeID, workingDirectory)
	case pxi.VolumeTypeZFS:
		return backupZFSVolume(sourcePath, volumeID, workingDirectory)
	case pxi.VolumeTypeRBD:
		return backupRBDVolume(sourcePath, volumeID, workingDirectory)
	case pxi.VolumeTypeLXCRootfs:
		return backupLXCRootfsVolume(sourcePath, volumeID, workingDirectory)
	default:
		return "", fmt.Errorf("unsupported volume type: %s", volumeType)
	}
}
