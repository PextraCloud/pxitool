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

type CompressionAlgorithm string

const (
	// CompressionAlgorithmZstd represents the Zstandard compression algorithm.
	CompressionAlgorithmZstd CompressionAlgorithm = "zstd"
	// CompressionAlgorithmGzip represents the Gzip compression algorithm.
	CompressionAlgorithmGzip CompressionAlgorithm = "gzip"
)

// CompressionAlgorithms is a list of supported compression algorithms.
var CompressionAlgorithms = []CompressionAlgorithm{CompressionAlgorithmZstd, CompressionAlgorithmGzip}

// CompressionAlgorithmsStrings is a list of supported compression algorithms as []string.
var CompressionAlgorithmsStrings = []string{CompressionAlgorithmZstd.String(), CompressionAlgorithmGzip.String()}

// Valid checks if the compression algorithm is valid.
func (c CompressionAlgorithm) Valid() bool {
	switch c {
	case CompressionAlgorithmZstd, CompressionAlgorithmGzip:
		return true
	default:
		return false
	}
}

func (c CompressionAlgorithm) String() string {
	return string(c)
}
