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
package dar

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os/exec"
)

// ListOptions holds the options for listing the contents of a dar archive. Only a subset of the options are supported.
type ListOptions struct {
	InputBasename string
}

// Args returns the command-line arguments for listing the contents of a dar archive with the specified options.
func (o ListOptions) Args() []string {
	args := []string{"-N", "-Q", "--list-format=xml", "-l", o.InputBasename}
	return args
}

/*
<?xml version="1.0" ?>
<!DOCTYPE Catalog SYSTEM "dar-catalog.dtd">
<Catalog format="1.2">
<File name="vol-DW0Mm2TL7kPcqSvCBs9tB.vmdk" size="192 kio" stored="513 o" crc="b4214416" dirty="no" sparse="yes" delta_sig="no" patch_base_crc="" patch_result_crc="">
<Attributes data="saved" metadata="absent" user="user" group="user" permissions=" -rw-r--r--" atime="1793416430" mtime="1793416430" ctime="1793416430" />
</File>
...
</Catalog>
*/
type listXMLFile struct {
	XMLName xml.Name `xml:"File"`
	Name    string   `xml:"name,attr"`
}
type listXML struct {
	XMLName xml.Name      `xml:"Catalog"`
	Files   []listXMLFile `xml:"File"`
}

// List lists the contents of a dar archive and returns the names of the files in the archive.
func List(options ListOptions) ([]string, error) {
	args := options.Args()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("dar", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list dar archive: %w", cmdError(err, stderr))
	}

	var catalog listXML
	if err := xml.Unmarshal(stdout.Bytes(), &catalog); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dar archive list: %w", err)
	}

	var names []string
	for _, file := range catalog.Files {
		names = append(names, file.Name)
	}
	return names, nil
}
