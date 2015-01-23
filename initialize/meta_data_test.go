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
	"net"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource"
)

func TestExtractIPsFromMetadata(t *testing.T) {
	for i, tt := range []struct {
		in  datasource.Metadata
		out map[string]string
	}{
		{
			datasource.Metadata{
				PublicIPv4:  net.ParseIP("12.34.56.78"),
				PrivateIPv4: net.ParseIP("1.2.3.4"),
				PublicIPv6:  net.ParseIP("1234::"),
				PrivateIPv6: net.ParseIP("5678::"),
			},
			map[string]string{"$public_ipv4": "12.34.56.78", "$private_ipv4": "1.2.3.4", "$public_ipv6": "1234::", "$private_ipv6": "5678::"},
		},
	} {
		got := ExtractIPsFromMetadata(tt.in)
		if !reflect.DeepEqual(got, tt.out) {
			t.Errorf("case %d: got %s, want %s", i, got, tt.out)
		}
	}
}
