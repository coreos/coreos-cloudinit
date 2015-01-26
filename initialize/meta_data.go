// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package initialize

import (
	"sort"

	"github.com/coreos/coreos-cloudinit/datasource"
)

// ExtractIPsFromMetaData parses a JSON blob in the OpenStack metadata service
// format and returns a substitution map possibly containing private_ipv4,
// public_ipv4, private_ipv6, and public_ipv6 addresses.
func ExtractIPsFromMetadata(metadata datasource.Metadata) map[string]string {
	subs := map[string]string{}
	if metadata.PrivateIPv4 != nil {
		subs["$private_ipv4"] = metadata.PrivateIPv4.String()
	}
	if metadata.PublicIPv4 != nil {
		subs["$public_ipv4"] = metadata.PublicIPv4.String()
	}
	if metadata.PrivateIPv6 != nil {
		subs["$private_ipv6"] = metadata.PrivateIPv6.String()
	}
	if metadata.PublicIPv6 != nil {
		subs["$public_ipv6"] = metadata.PublicIPv6.String()
	}

	return subs
}

func sortedKeys(m map[string]string) (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}
