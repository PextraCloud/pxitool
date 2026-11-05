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
	"os"
	"os/exec"
	"strings"
)

// CreateOptions holds the options for creating a dar archive. Only a subset of the options are supported.
type CreateOptions struct {
	OutputBasename string
	RootDir        string
	Overwrite      bool
	Files          []string
	DryRun         bool
	Cipher         string
	Key            string
	Compression    string
}

// Args returns the command-line arguments for creating a dar archive with the specified options.
func (o CreateOptions) Args() []string {
	args := []string{"-N", "-Q", "-c"}
	if o.DryRun {
		args = append(args, "-e")
	}
	args = append(args, o.OutputBasename, "-R", o.RootDir)

	for _, file := range o.Files {
		args = append(args, "-g", file)
	}
	if o.Cipher != "" || o.Key != "" {
		// TODO: create a file instead of passing through cmdline
		args = append(args, "-K", fmt.Sprintf("%s:%s", o.Cipher, o.Key))
	}
	if o.Compression != "" {
		args = append(args, "--compression", o.Compression)
	}
	if !o.Overwrite {
		args = append(args, "-n")
	} else {
		args = append(args, "-w")
	}
	return args
}

// Create creates a dar archive with the specified options. It returns an error if the creation fails.
//
// The following default behaviors apply:
//
// - Only options.Files are included in the archive
// - dar is instructed not to read dar config files
func Create(options CreateOptions) error {
	if options.Key != "" && options.Cipher == "" {
		return fmt.Errorf("key specified without cipher")
	} else if options.Key == "" && options.Cipher != "" {
		return fmt.Errorf("cipher specified without key")
	}
	options.OutputBasename = strings.TrimSpace(options.OutputBasename)

	args := options.Args()
	cmd := exec.Command("dar", args...)
	if options.OutputBasename == "-" {
		// Write to stdout
		cmd.Stdout = os.Stdout
	}
	return cmd.Run()
}
