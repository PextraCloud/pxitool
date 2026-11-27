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
package pxi

// VolumeType represents the type of a storage volume. Constrained from Pextra CloudEnvironment's StoragePoolType enum.
type VolumeType string

const (
	// VolumeTypeDirectory represents a volume stored in a filesystem directory (e.g., a QEMU image file).
	VolumeTypeDirectory VolumeType = "directory"
	// VolumeTypeLVM represents a LVM volume.
	VolumeTypeLVM VolumeType = "lvm"
	// VolumeTypeZFS represents a ZFS volume.
	VolumeTypeZFS VolumeType = "zfs"
	// VolumeTypeRBD represents a RBD volume.
	VolumeTypeRBD VolumeType = "rbd"
	// VolumeTypeLXCRootfs represents a LXC rootfs stored in a filesystem.
	VolumeTypeLXCRootfs VolumeType = "lxc_rootfs"
)

// VolumeTypes is a list of supported volume types.
var VolumeTypes = []VolumeType{VolumeTypeDirectory, VolumeTypeLVM, VolumeTypeZFS, VolumeTypeRBD, VolumeTypeLXCRootfs}

// Valid checks if the volume type is valid.
func (c VolumeType) Valid() bool {
	switch c {
	case VolumeTypeDirectory, VolumeTypeLVM, VolumeTypeZFS, VolumeTypeRBD, VolumeTypeLXCRootfs:
		return true
	default:
		return false
	}
}

func (c VolumeType) String() string {
	return string(c)
}
