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
	"os"

	"github.com/PextraCloud/pxitool/pkg/log"
	"github.com/spf13/cobra"
)

var debugFlag bool

var rootCmd = &cobra.Command{
	Use:   "pxitool",
	Short: "CLI tool for working with .pxi files (Pextra Images)",
	Long: `Pextra Image Tool (pxitool) is a command-line utility
for creating, extracting, and managing Pextra Images,
or .pxi files. It fully conforms to specification at
https://pextracloud.github.io/ImageFormats/pxi.

Copyright (C) 2025 Pextra Inc. This tool is licensed
under the Apache License, Version 2.0.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetDebug(debugFlag)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
