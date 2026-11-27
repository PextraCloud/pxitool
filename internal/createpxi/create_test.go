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
package createpxi

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/PextraCloud/pxitool/internal/testutils"
	"github.com/PextraCloud/pxitool/pkg/pxi"
)

func setupDar(t *testing.T) (logFile string) {
	t.Helper()
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	if err := os.Mkdir(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	logFile = filepath.Join(tmp, "dar-args.log")
	testutils.StubDarLog(t, binDir, logFile)
	return logFile
}

func TestCreatePXIFunction(t *testing.T) {
	tests := []struct {
		name          string
		options       CreatePXIOptions
		wantTokens    []string
		wantNotTokens []string
	}{
		{
			name: "aes256 + zstd",
			options: CreatePXIOptions{
				EncryptionCipher:     pxi.EncryptionCipherAES256,
				EncryptionKey:        "secret",
				CompressionAlgorithm: pxi.CompressionAlgorithmZstd,
				Overwrite:            true,
			},
			wantTokens:    []string{"aes:secret", "zstd", "-w"},
			wantNotTokens: []string{"gzip", "-n"},
		},
		{
			name: "aes256 + gzip",
			options: CreatePXIOptions{
				EncryptionCipher:     pxi.EncryptionCipherAES256,
				EncryptionKey:        "secret",
				CompressionAlgorithm: pxi.CompressionAlgorithmGzip,
			},
			wantTokens:    []string{"aes:secret", "gzip", "-n"},
			wantNotTokens: []string{"zstd", "-w"},
		},
		{
			name:          "no cipher or compression",
			options:       CreatePXIOptions{},
			wantNotTokens: []string{"-K", "aes:", "--compression", "zstd", "gzip"},
		},
		{
			name: "unknown cipher and compression map to empty",
			options: CreatePXIOptions{
				EncryptionCipher:     pxi.EncryptionCipher("chacha20"),
				CompressionAlgorithm: pxi.CompressionAlgorithm("lz4"),
			},
			wantNotTokens: []string{"-K", "--compression", "chacha20", "lz4", "zstd", "gzip", "-w"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logFile := setupDar(t)
			opts := tt.options
			opts.OutputPath = filepath.Join(t.TempDir(), "out.pxi")
			err := createPXI(opts, t.TempDir(), []string{pxi.ManifestFileName})
			if err != nil {
				t.Fatalf("createPXI() error = %v", err)
			}
			tokens := argsTokens(t, logFile)
			for _, want := range tt.wantTokens {
				if !slices.Contains(tokens, want) {
					t.Errorf("dar args %v missing %q", tokens, want)
				}
			}
			for _, notWant := range tt.wantNotTokens {
				if slices.Contains(tokens, notWant) {
					t.Errorf("dar args %v should not contain %q", tokens, notWant)
				}
			}
		})
	}
}

func TestCreatePXIErrorCases(t *testing.T) {
	t.Run("key without valid cipher", func(t *testing.T) {
		setupDar(t)
		opts := CreatePXIOptions{
			OutputPath:       filepath.Join(t.TempDir(), "out.pxi"),
			EncryptionCipher: pxi.EncryptionCipher("chacha20"),
			EncryptionKey:    "secret",
		}
		err := createPXI(opts, t.TempDir(), []string{pxi.ManifestFileName})
		if err == nil || !strings.Contains(err.Error(), "key specified without cipher") {
			t.Fatalf("createPXI() error = %v, want key specified without cipher", err)
		}
	})

	t.Run("cipher without key", func(t *testing.T) {
		setupDar(t)
		opts := CreatePXIOptions{
			OutputPath:       filepath.Join(t.TempDir(), "out.pxi"),
			EncryptionCipher: pxi.EncryptionCipherAES256,
		}
		err := createPXI(opts, t.TempDir(), []string{pxi.ManifestFileName})
		if err == nil || !strings.Contains(err.Error(), "cipher specified without key") {
			t.Fatalf("createPXI() error = %v, want cipher specified without key", err)
		}
	})

	t.Run("dar failure", func(t *testing.T) {
		tmp := t.TempDir()
		binDir := filepath.Join(tmp, "bin")
		if err := os.Mkdir(binDir, 0755); err != nil {
			t.Fatal(err)
		}
		testutils.StubExit(t, "dar", 3)

		opts := CreatePXIOptions{
			OutputPath: filepath.Join(t.TempDir(), "out.pxi"),
		}
		err := createPXI(opts, t.TempDir(), []string{pxi.ManifestFileName})
		if err == nil {
			t.Fatal("createPXI() = nil, want error")
		}
	})
}
