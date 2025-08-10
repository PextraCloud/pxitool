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
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSetDebugAndIsDebug(t *testing.T) {
	originalState := IsDebug()
	t.Cleanup(func() {
		SetDebug(originalState)
	})

	SetDebug(true)
	if !IsDebug() {
		t.Error("expected IsDebug() to be true after SetDebug(true)")
	}

	SetDebug(false)
	if IsDebug() {
		t.Error("expected IsDebug() to be false after SetDebug(false)")
	}
}

func TestLoggingFunctions(t *testing.T) {
	originalOutput := output
	var buf bytes.Buffer
	SetOutput(&buf)
	t.Cleanup(func() {
		SetOutput(originalOutput)
	})

	originalDebugState := IsDebug()
	t.Cleanup(func() {
		SetDebug(originalDebugState)
	})

	testCases := []struct {
		name           string
		logFunc        func(format string, args ...any)
		debugEnabled   bool
		expectedPrefix string
		shouldLog      bool
	}{
		{"Info", Info, false, "[INFO]", true},
		{"Warn", Warn, false, "[WARN]", true},
		{"Error", Error, false, "[ERROR]", true},
		{"Debug when disabled", Debug, false, "", false},
		{"Debug when enabled", Debug, true, "[DEBUG]", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			SetDebug(tc.debugEnabled)

			tc.logFunc("test message %d", 123)
			logOutput := buf.String()

			if tc.shouldLog {
				if !strings.HasPrefix(logOutput, tc.expectedPrefix) {
					t.Errorf("expected prefix %q, got %q", tc.expectedPrefix, logOutput)
				}
				if !strings.Contains(logOutput, "test message 123") {
					t.Errorf("expected log to contain message, but it did not. Got: %q", logOutput)
				}
			} else {
				if logOutput != "" {
					t.Errorf("expected no log output, but got: %q", logOutput)
				}
			}
		})
	}
}

func TestSetOutput(t *testing.T) {
	originalOutput := output
	t.Cleanup(func() {
		output = originalOutput
	})

	// Test setting a new writer
	var buf bytes.Buffer
	SetOutput(&buf)
	if output != &buf {
		t.Error("SetOutput did not set the new writer")
	}

	// Test logging to the new writer
	Info("hello")
	if !strings.Contains(buf.String(), "[INFO] hello") {
		t.Errorf("log was not written to the new output buffer. Got: %q", buf.String())
	}

	// Test resetting to default with nil
	SetOutput(nil)
	if output != os.Stderr {
		t.Error("SetOutput(nil) did not reset output to os.Stderr")
	}
}
