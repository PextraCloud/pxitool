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
	"bytes"

	"github.com/PextraCloud/pxitool/pkg/pxi/constants/volumeformat"
)

var (
	QCOW2Signature = []byte{0x51, 0x46, 0x49, 0xfb} // QCOW2 signature ("magic number")
	VMDKSignature  = []byte{0x4b, 0x44, 0x4d, 0x56} // VMware VMDK signature ("magic number")
)

func getVolumeFormat(data [4]byte) volumeformat.VolumeFormat {
	switch {
	case bytes.Equal(data[:], QCOW2Signature):
		return volumeformat.QCOW2
	case bytes.Equal(data[:], VMDKSignature):
		return volumeformat.VMDK
	default:
		// Default to Raw if no known signature matches
		return volumeformat.Raw
	}
}
