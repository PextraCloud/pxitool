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
	"reflect"
	"testing"
)

func TestCreateOptionsArgs(t *testing.T) {
	tests := []struct {
		name     string
		options  CreateOptions
		expected []string
	}{
		{
			name: "minimal options",
			options: CreateOptions{
				OutputBasename: "out.dar",
				RootDir:        "/root",
			},
			expected: []string{"-N", "-Q", "-c", "out.dar", "-R", "/root", "-n"},
		},
		{
			name: "dry run",
			options: CreateOptions{
				OutputBasename: "out.dar",
				RootDir:        "/root",
				DryRun:         true,
			},
			expected: []string{"-N", "-Q", "-c", "-e", "out.dar", "-R", "/root", "-n"},
		},
		{
			name: "files",
			options: CreateOptions{
				OutputBasename: "out.dar",
				RootDir:        "/root",
				Files:          []string{"a.txt", "b.txt"},
			},
			expected: []string{"-N", "-Q", "-c", "out.dar", "-R", "/root", "-g", "a.txt", "-g", "b.txt", "-n"},
		},
		{
			name: "cipher and key",
			options: CreateOptions{
				OutputBasename: "out.dar",
				RootDir:        "/root",
				Cipher:         "aes",
				Key:            "secret",
			},
			expected: []string{"-N", "-Q", "-c", "out.dar", "-R", "/root", "-K", "aes:secret", "-n"},
		},
		{
			name: "compression",
			options: CreateOptions{
				OutputBasename: "out.dar",
				RootDir:        "/root",
				Compression:    "zstd",
			},
			expected: []string{"-N", "-Q", "-c", "out.dar", "-R", "/root", "--compression", "zstd", "-n"},
		},
		{
			name: "overwrite",
			options: CreateOptions{
				OutputBasename: "out.dar",
				RootDir:        "/root",
				Overwrite:      true,
			},
			expected: []string{"-N", "-Q", "-c", "out.dar", "-R", "/root", "-w"},
		},
		{
			name: "all options",
			options: CreateOptions{
				OutputBasename: "out.dar",
				RootDir:        "/root",
				Overwrite:      true,
				Files:          []string{"a.txt"},
				DryRun:         true,
				Cipher:         "aes",
				Key:            "secret",
				Compression:    "zstd",
			},
			expected: []string{
				"-N", "-Q", "-c", "-e", "out.dar", "-R", "/root",
				"-g", "a.txt",
				"-K", "aes:secret",
				"--compression", "zstd",
				"-w",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.options.Args()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Args() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractOptionsArgs(t *testing.T) {
	tests := []struct {
		name     string
		options  ExtractOptions
		expected []string
	}{
		{
			name: "minimal options",
			options: ExtractOptions{
				InputBasename: "in.dar",
				RootDir:       "/root",
			},
			expected: []string{"-N", "-Q", "--comparison-field=mtime", "-x", "in.dar", "-R", "/root", "-n"},
		},
		{
			name: "dry run",
			options: ExtractOptions{
				InputBasename: "in.dar",
				RootDir:       "/root",
				DryRun:        true,
			},
			expected: []string{"-N", "-Q", "--comparison-field=mtime", "-x", "-e", "in.dar", "-R", "/root", "-n"},
		},
		{
			name: "files",
			options: ExtractOptions{
				InputBasename: "in.dar",
				RootDir:       "/root",
				Files:         []string{"a.txt", "b.txt"},
			},
			expected: []string{"-N", "-Q", "--comparison-field=mtime", "-x", "in.dar", "-R", "/root", "-g", "a.txt", "-g", "b.txt", "-n"},
		},
		{
			name: "cipher and key",
			options: ExtractOptions{
				InputBasename: "in.dar",
				RootDir:       "/root",
				Cipher:        "aes",
				Key:           "secret",
			},
			expected: []string{"-N", "-Q", "--comparison-field=mtime", "-x", "in.dar", "-R", "/root", "-K", "aes:secret", "-n"},
		},
		{
			name: "overwrite",
			options: ExtractOptions{
				InputBasename: "in.dar",
				RootDir:       "/root",
				Overwrite:     true,
			},
			expected: []string{"-N", "-Q", "--comparison-field=mtime", "-x", "in.dar", "-R", "/root", "-w"},
		},
		{
			name: "all options",
			options: ExtractOptions{
				InputBasename: "in.dar",
				RootDir:       "/root",
				Overwrite:     true,
				Files:         []string{"a.txt"},
				DryRun:        true,
				Cipher:        "aes",
				Key:           "secret",
			},
			expected: []string{
				"-N", "-Q", "--comparison-field=mtime", "-x", "-e", "in.dar", "-R", "/root",
				"-g", "a.txt",
				"-K", "aes:secret",
				"-w",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.options.Args()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Args() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestListOptionsArgs(t *testing.T) {
	tests := []struct {
		name     string
		options  ListOptions
		expected []string
	}{
		{
			name:     "input basename",
			options:  ListOptions{InputBasename: "in.dar"},
			expected: []string{"-N", "-Q", "-l", "in.dar", "-aheader"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.options.Args()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Args() = %v, want %v", got, tt.expected)
			}
		})
	}
}
