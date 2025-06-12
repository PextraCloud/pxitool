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
package readpxi

import (
	"io"

	"github.com/PextraCloud/pxitool/pkg/pxi/signature"
)

func verifySignature(reader io.Reader) error {
	magic := make([]byte, signature.PXISignatureLength)
	if _, err := reader.Read(magic); err != nil {
		return err
	}
	if err := signature.Verify(magic); err != nil {
		return err
	}

	return nil
}
