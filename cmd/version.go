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

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of pxitool",
	Long: `Prints the current version of the Pextra Image
Tool (pxitool), along with the version of the
Pextra Image Specification it conforms to.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Pextra Image Tool version 1.0.0")
		cmd.Println("Conforms to Pextra Image Specification version 1.0")
	},
}
