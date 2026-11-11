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
package version

import (
	"fmt"

	"github.com/PextraCloud/pxitool/internal/dar"
)

// PxiToolVersion is the version string for pxitool.
const PxiToolVersion = "2.0.0"

// Version returns the version string for pxitool, including the dar version if available.
func Version() string {
	darVersion := "unknown (error retrieving dar version)"
	darVersionInfo, err := dar.Version()
	if err == nil {
		darVersion = fmt.Sprintf("%s, %s", darVersionInfo.DarVersion, darVersionInfo.LibDarVersion)
	}

	return fmt.Sprintf("pxitool %s (%s)", PxiToolVersion, darVersion)
}
