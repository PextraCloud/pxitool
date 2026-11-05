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
)

// ExtractOptions holds the options for extracting a dar archive. Only a subset of the options are supported.
type ExtractOptions struct {
	InputBasename string
	RootDir       string
	Overwrite     bool
	Files         []string
	DryRun        bool
	Cipher        string
	Key           string
}

// Args returns the command-line arguments for extracting a dar archive with the specified options.
func (o ExtractOptions) Args() []string {
	args := []string{"-N", "-Q", "--comparison-field=mtime", "-x"}
	if o.DryRun {
		args = append(args, "-e")
	}
	args = append(args, o.InputBasename, "-R", o.RootDir)

	if len(o.Files) > 0 {
		for _, file := range o.Files {
			args = append(args, "-g", file)
		}
	}
	if o.Cipher != "" || o.Key != "" {
		// TODO: create a file instead of passing through cmdline
		args = append(args, "-K", fmt.Sprintf("%s:%s", o.Cipher, o.Key))
	}
	if !o.Overwrite {
		args = append(args, "-n")
	} else {
		args = append(args, "-w")
	}

	return args
}

// Extract extracts files from a dar archive with the specified options. It returns an error if the extraction fails.
//
// The following default behaviors apply:
//
// - Only options.Files are extracted from the archive
// - dar is instructed not to read dar config files
// - dar is instructed to ignore file ownership and permissions, only considering modification time (mtime)
func Extract(options ExtractOptions) error {
	args := options.Args()
	cmd := exec.Command("dar", args...)
	return cmd.Run()
}
