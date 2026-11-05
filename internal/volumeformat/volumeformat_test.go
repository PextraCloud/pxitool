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
package volumeformat

import (
	"testing"
)

func hdr(s string) [4]byte {
	var b [4]byte
	copy(b[:], s)
	return b
}

func TestGetVolumeFormat(t *testing.T) {
	for _, tc := range []struct {
		hdr  [4]byte
		want VolumeFormat
	}{
		{hdr(string(QCOW2Signature)), VolumeFormatQCOW2},
		{hdr(string(VMDKSignature)), VolumeFormatVMDK},
		// Raw / unknown samples.
		{hdr("PK\x03\x04"), VolumeFormatRaw}, // zip magic (F5)
		{hdr("QFI\xfc"), VolumeFormatRaw},    // qcow2 magic + 1 byte
		{hdr("QFG\xfb"), VolumeFormatRaw},    // qcow2 magic, 2nd byte off
		{hdr("KDMW"), VolumeFormatRaw},       // vmdk magic + 1 byte
		{hdr("KDMS"), VolumeFormatRaw},       // vmdk magic, 2nd byte off
		{hdr("\x00\x00\x00\x00"), VolumeFormatRaw},
		{hdr("\xff\xff\xff\xff"), VolumeFormatRaw},
		{hdr("\x01\x02\x03\x04"), VolumeFormatRaw},
		{hdr("\x00\x01\x02\x03"), VolumeFormatRaw},
		{hdr("raw!"), VolumeFormatRaw},
		{hdr("\xfb\x49\x46\x51"), VolumeFormatRaw}, // qcow2 magic reversed
		{hdr("\x56\x4d\x44\x4b"), VolumeFormatRaw}, // vmdk magic reversed
		{hdr("\x01\x00\x00\x00"), VolumeFormatRaw},
		{hdr("\x00\x00\x00\x01"), VolumeFormatRaw},
	} {
		if got := GetVolumeFormat(tc.hdr); got != tc.want {
			t.Errorf("GetVolumeFormat(% x) = %v, want %v", tc.hdr[:], got, tc.want)
		}
	}
}
