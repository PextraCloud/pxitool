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
package signature

import (
	"bytes"
	"errors"
)

var (
	PXISignature        = []byte{0x50, 0x58, 0x49, 0x00, 0x0D, 0x0A, 0x1A, 0x0A} // The PXI signature ("magic number")
	PXISignatureLength  = len(PXISignature)
	ErrInvalidSignature = errors.New("invalid PXI signature")
)

// Checks if the provided byte slice has the PXI signature.
func Verify(data []byte) error {
	if len(data) < len(PXISignature) {
		return ErrInvalidSignature
	}
	if !bytes.Equal(data[:len(PXISignature)], PXISignature) {
		return ErrInvalidSignature
	}
	return nil
}
