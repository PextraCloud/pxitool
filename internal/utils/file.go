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

import (
	"fmt"
	"os"
	"strings"
)

func promptOverwrite(fileName string) bool {
	var response string
	fmt.Printf("Output file %s already exists. Overwrite? (y/N): ", fileName)
	if _, err := fmt.Scanln(&response); err != nil {
		fmt.Println("Error reading input:", err)
		return false
	}
	response = strings.TrimSpace(response)
	if strings.EqualFold(response, "y") || strings.EqualFold(response, "yes") {
		return true
	} else {
		fmt.Println("Operation cancelled.")
		return false
	}
}

func GetOutputFileHandle(name string, force bool) (*os.File, error) {
	flags := os.O_WRONLY | os.O_CREATE
	// If force, truncate file, otherwise use exclusive creation
	if force {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	file, err := os.OpenFile(name, flags, 0644)
	if err != nil {
		// If file exists and not forcing, prompt user to overwrite
		if os.IsExist(err) && !force && promptOverwrite(name) {
			return GetOutputFileHandle(name, true)
		}
		return nil, err
	}
	return file, nil
}
