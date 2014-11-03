/*
   Copyright 2014 CoreOS, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package initialize

import (
	"fmt"
	"log"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"
)

func ParseUserData(contents string) (interface{}, error) {
	if len(contents) == 0 {
		return nil, nil
	}
	header := strings.SplitN(contents, "\n", 2)[0]

	// Explicitly trim the header so we can handle user-data from
	// non-unix operating systems. The rest of the file is parsed
	// by yaml, which correctly handles CRLF.
	header = strings.TrimSpace(header)

	if strings.HasPrefix(header, "#!") {
		log.Printf("Parsing user-data as script")
		return system.Script(contents), nil
	} else if header == "#cloud-config" {
		log.Printf("Parsing user-data as cloud-config")
		return NewCloudConfig(contents)
	} else {
		return nil, fmt.Errorf("Unrecognized user-data header: %s", header)
	}
}
