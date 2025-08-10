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
package backup

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestBackupDockerVolume(t *testing.T) {
	var w io.Writer = &bytes.Buffer{}
	err := BackupDockerVolume("any_docker_volume", &w)

	if err == nil {
		t.Fatal("Expected an error from BackupDockerVolume, but got nil")
	}

	if !strings.Contains(err.Error(), "not supported yet") {
		t.Errorf("Expected error message to contain 'not supported yet', but got: %v", err)
	}
}
