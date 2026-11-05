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
package volumeformat

import (
	"bytes"
)

var (
	// QCOW2Signature is the signature (magic number) for QCOW2 disk images.
	QCOW2Signature = []byte{0x51, 0x46, 0x49, 0xfb}
	// VMDKSignature is the signature (magic number) for VMware VMDK disk images.
	VMDKSignature = []byte{0x4b, 0x44, 0x4d, 0x56}
)

// VolumeFormat represents the format of a disk volume.
type VolumeFormat string

const (
	// VolumeFormatRaw represents a raw disk image format.
	VolumeFormatRaw VolumeFormat = "raw"
	// VolumeFormatQCOW2 represents the QCOW2 disk image format.
	VolumeFormatQCOW2 VolumeFormat = "qcow2"
	// VolumeFormatVMDK represents the VMware VMDK disk image format.
	VolumeFormatVMDK VolumeFormat = "vmdk"
)

// GetVolumeFormat determines the volume format based on the provided 4-byte signature.
func GetVolumeFormat(data [4]byte) VolumeFormat {
	switch {
	case bytes.Equal(data[:], QCOW2Signature):
		return VolumeFormatQCOW2
	case bytes.Equal(data[:], VMDKSignature):
		return VolumeFormatVMDK
	default:
		// Default to Raw if no known signature matches
		return VolumeFormatRaw
	}
}

func (v VolumeFormat) String() string {
	switch v {
	case VolumeFormatRaw:
		return "raw"
	case VolumeFormatQCOW2:
		return "qcow2"
	case VolumeFormatVMDK:
		return "vmdk"
	default:
		panic("unknown volume format")
	}
}
