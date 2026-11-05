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
	"os/exec"
	"regexp"
	"strings"
)

var (
	darVersionRegex    = regexp.MustCompile(`^\s+dar.*(\d+\.\d+\.\d+)`)
	libdarVersionRegex = regexp.MustCompile(`libdar.*(\d+\.\d+\.\d+)`)
)

// VersionInfo holds the version information of dar and libdar.
type VersionInfo struct {
	DarVersion    string
	LibDarVersion string
}

// Version returns the version information of dar and libdar.
func Version() (*VersionInfo, error) {
	outputByte, err := exec.Command("dar", "--version").CombinedOutput()
	if err != nil {
		return nil, err
	}

	output := string(outputByte)

	darVersionMatches := darVersionRegex.FindAllString(output, -1)
	libdarVersionMatches := libdarVersionRegex.FindAllString(output, -1)

	darVersion := strings.TrimSpace(strings.Join(darVersionMatches, "\n"))
	libdarVersion := strings.TrimSpace(strings.Join(libdarVersionMatches, "\n"))
	return &VersionInfo{
		DarVersion:    darVersion,
		LibDarVersion: libdarVersion,
	}, nil
}
