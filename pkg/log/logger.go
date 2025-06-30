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
package log

import (
	"fmt"
	"io"
	"os"
)

var (
	debugEnabled bool
	output       io.Writer = os.Stderr
)

// SetDebug enables or disables debug logging
func SetDebug(enabled bool) {
	debugEnabled = enabled
}

// Debug prints debug messages if debug is enabled
func Debug(format string, args ...any) {
	if debugEnabled {
		fmt.Fprintf(output, "[DEBUG] "+format+"\n", args...)
	}
}

// Info prints informational messages
func Info(format string, args ...any) {
	fmt.Fprintf(output, "[INFO] "+format+"\n", args...)
}

// Warn prints warning messages
func Warn(format string, args ...any) {
	fmt.Fprintf(output, "[WARN] "+format+"\n", args...)
}

// Error prints error messages
func Error(format string, args ...any) {
	fmt.Fprintf(output, "[ERROR] "+format+"\n", args...)
}
