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
package dar

import (
	"fmt"
	"os/exec"
	"regexp"
)

var (
	darSymmetricEncryptionRegex  = regexp.MustCompile(`Symmetric key encryption used\s*:\s*(\S+)`)
	darCompressionAlgorithmRegex = regexp.MustCompile(`Compression algorithm used\s*:\s*(\S+)`)
)

type Header struct {
	CompressionAlgorithm         string `json:"compression_algorithm"`
	SymmetricEncryptionAlgorithm string `json:"symmetric_encryption_algorithm"`
	// IsEncrypted indicates whether the archive is encrypted or not. Determined by checking if
	// SymmetricEncryptionAlgorithm is not "none".
	IsEncrypted bool `json:"is_encrypted"`
}

// ListOptions holds the options for listing the contents of a dar archive. Only a subset of the options are supported.
type ListOptions struct {
	InputBasename string
}

// Args returns the command-line arguments for listing the contents of a dar archive with the specified options.
func (o ListOptions) Args() []string {
	args := []string{"-N", "-Q", "-l"}
	args = append(args, o.InputBasename, "-aheader")
	return args
}

// parseHeader parses the output of the dar command to extract the header information, including compression algorithm
// and symmetric encryption algorithm.
func parseHeader(output string) (*Header, error) {
	header := &Header{}

	/*
		Archive version format               : 11.3
		Compression algorithm used           : zstd
		Compression block size used          : 0
		Symmetric key encryption used        : none
		Asymmetric key encryption used       : none
		Archive is signed                    : no
		Sequential reading marks             : present
		User comment                         : N/A
	*/

	// Extract the symmetric encryption algorithm
	symmetricEncryptionMatches := darSymmetricEncryptionRegex.FindStringSubmatch(output)
	if len(symmetricEncryptionMatches) < 2 {
		return nil, fmt.Errorf("failed to parse symmetric encryption algorithm from dar output")
	}
	header.SymmetricEncryptionAlgorithm = symmetricEncryptionMatches[1]
	// Set IsEncrypted
	header.IsEncrypted = header.SymmetricEncryptionAlgorithm != "none"

	// Extract the compression algorithm
	compressionAlgorithmMatches := darCompressionAlgorithmRegex.FindStringSubmatch(output)
	if len(compressionAlgorithmMatches) < 2 {
		return nil, fmt.Errorf("failed to parse compression algorithm from dar output")
	}
	header.CompressionAlgorithm = compressionAlgorithmMatches[1]

	return header, nil
}

// ListHeader reads the header of a dar archive and returns the compression algorithm and symmetric encryption algorithm
// used in the archive.
func ListHeader(options ListOptions) (*Header, error) {
	args := options.Args()

	// NOTE: dar exits with a non-zero exit code with -aheader
	cmd := exec.Command("dar", args...)
	outputByte, _ := cmd.CombinedOutput()
	output := string(outputByte)

	return parseHeader(output)
}

func (h *Header) String() string {
	return fmt.Sprintf("Compression Algorithm: %s\nSymmetric Encryption Algorithm: %s", h.CompressionAlgorithm, h.SymmetricEncryptionAlgorithm)
}
