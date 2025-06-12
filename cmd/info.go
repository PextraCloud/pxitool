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
	"encoding/json"
	"fmt"
	"os"

	"github.com/PextraCloud/pxitool/internal/readpxi"
	"github.com/spf13/cobra"
)

var isMetadataDump bool

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().BoolVarP(&isMetadataDump, "metadata", "m", false, "Dump PXI metadata in JSON format")
}

var infoCmd = &cobra.Command{
	Use:   "info [file]",
	Args:  cobra.ExactArgs(1),
	Short: "Display information about a Pextra Image",
	Long: `This command displays information about a
Pextra Image (.pxi) file. It checks the file structure,
verifies the integrity of the chunks, and outputs the
metadata contained within the image.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFileName := args[0]
		result, err := readpxi.Read(inputFileName)
		if err != nil {
			cmd.PrintErrf("Error reading PXI file: %v\n", err)
			os.Exit(1)
		}

		if isMetadataDump {
			var jsonData []byte
			if jsonData, err = json.MarshalIndent(result.CONF.Config, "", "    "); err != nil {
				cmd.PrintErrf("Error serializing metadata to JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonData))
		}
	},
}
