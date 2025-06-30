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
package utils

import "github.com/PextraCloud/pxitool/pkg/pxi/chunks/conf"

// Extracts volume paths from the configuration, excluding specified IDs.
func GetVolumePathsFromConfig(config *conf.InstanceConfigGeneric, excluded []string) []string {
	excludedSet := make(map[string]struct{}, len(excluded))
	for _, id := range excluded {
		excludedSet[id] = struct{}{}
	}

	var paths []string
	for _, v := range config.Volumes {
		if _, found := excludedSet[v.ID]; !found {
			paths = append(paths, v.Path)
		}
	}
	return paths
}
