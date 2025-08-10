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
package cmd

import (
	"bytes"
	"testing"

	"github.com/PextraCloud/pxitool/pkg/log"
)

func TestRootCommand_DebugFlag(t *testing.T) {
	initialDebugState := log.IsDebug()
	log.SetDebug(false)
	t.Cleanup(func() {
		log.SetDebug(initialDebugState)
	})

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"version", "--debug"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if !log.IsDebug() {
		t.Error("expected debug flag to be set, but it was not")
	}
}
