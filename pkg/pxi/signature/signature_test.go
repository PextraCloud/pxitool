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

import "testing"

func TestVerify(t *testing.T) {
	correctSignatureWithExtra := append(PXISignature, []byte("extra data")...)
	incorrectSignature := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xDE, 0xAD, 0xBE, 0xEF}
	shortSignature := []byte{0x50, 0x58, 0x49}

	testCases := []struct {
		name          string
		input         []byte
		expectedError error
	}{
		{
			name:          "correct signature",
			input:         PXISignature,
			expectedError: nil,
		},
		{
			name:          "correct signature with extra data",
			input:         correctSignatureWithExtra,
			expectedError: nil,
		},
		{
			name:          "incorrect signature",
			input:         incorrectSignature,
			expectedError: ErrInvalidSignature,
		},
		{
			name:          "data too short",
			input:         shortSignature,
			expectedError: ErrInvalidSignature,
		},
		{
			name:          "nil data",
			input:         nil,
			expectedError: ErrInvalidSignature,
		},
		{
			name:          "empty data",
			input:         []byte{},
			expectedError: ErrInvalidSignature,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Verify(tc.input)
			if err != tc.expectedError {
				t.Errorf("expected error %v, but got %v", tc.expectedError, err)
			}
		})
	}
}
