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
package cli

import (
	"context"
	"fmt"

	"github.com/PextraCloud/pxitool/internal/version"
	"github.com/spf13/cobra"
)

var rootVersionFlag bool

const Version = "2.0.0"

var rootCmd = &cobra.Command{
	Use:          "pxitool",
	Short:        "CLI tool for working with .pxi files (Pextra Images)",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set it to log to stderr by default
	},
	Run: func(cmd *cobra.Command, args []string) {
		if !rootVersionFlag {
			cmd.Help()
			return
		}

		v := version.Version()
		fmt.Fprintln(cmd.OutOrStdout(), v)
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&rootVersionFlag, "version", "v", false, "Show version information")
}

func Execute(ctx context.Context) int {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return 1
	}
	return 0
}
