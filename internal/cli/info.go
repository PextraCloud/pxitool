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
	"encoding/json"
	"fmt"

	"github.com/PextraCloud/pxitool/internal/createpxi"
	"github.com/PextraCloud/pxitool/internal/readpxi"
	"github.com/spf13/cobra"
)

var infoJsonOutput bool
var infoEncryptionKey string
var infoOnlyHeader bool

var infoCmd = &cobra.Command{
	Use:     "info <file>",
	Args:    cobra.ExactArgs(1),
	Example: "  pxitool info /path/to/pxi/file",
	Short:   "Display information about a Pextra Image (PXI)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := createpxi.CheckEnvironment(); err != nil {
			return err
		}

		pxiPath := args[0]

		var toPrint fmt.Stringer
		var err error
		if infoOnlyHeader {
			toPrint, err = readpxi.ReadPXIHeader(pxiPath)
		} else {
			toPrint, err = readpxi.ReadPXI(pxiPath, infoEncryptionKey)
		}
		if err != nil {
			return err
		}

		if infoJsonOutput {
			output, err := json.MarshalIndent(toPrint, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(output))
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), toPrint.String())
		}
		return nil
	},
}

func init() {
	infoCmd.Flags().BoolVarP(&infoJsonOutput, "json", "j", false, "Output information in JSON format")
	infoCmd.Flags().StringVarP(&infoEncryptionKey, "key", "k", "", "Encryption key for the PXI file")
	infoCmd.Flags().BoolVarP(&infoOnlyHeader, "header", "H", false, "Only display the header information of the PXI file; encryption key is ignored")
	rootCmd.AddCommand(infoCmd)
}
