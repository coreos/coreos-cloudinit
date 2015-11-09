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

package config

import (
	"reflect"
	"testing"
)

func TestUserMerge(t *testing.T) {
	tests := []struct {
		inputs   []User
		expected User
	}{
		{
			inputs: []User{
				User{Name: "user", SSHAuthorizedKeys: []string{"key1"}},
				User{Name: "user", SSHAuthorizedKeys: []string{"key1", "key2"}},
			},
			expected: User{Name: "user", SSHAuthorizedKeys: []string{"key1", "key2"}},
		},
		{
			inputs: []User{
				User{Name: "user", SSHImportGithubUsers: []string{"github1"}},
				User{Name: "user", SSHImportGithubUsers: []string{"github1", "github2"}},
			},
			expected: User{Name: "user", SSHImportGithubUsers: []string{"github1", "github2"}},
		},
		{
			inputs: []User{
				User{Name: "user", Groups: []string{"group1"}},
				User{Name: "user", Groups: []string{"group1", "group2"}},
			},
			expected: User{Name: "user", Groups: []string{"group1", "group2"}},
		},
	}
	for i, tt := range tests {
		usr := tt.inputs[0]
		usr.Merge(tt.inputs[1])
		if !reflect.DeepEqual(tt.expected, usr) {
			t.Errorf("bad user (test case #%d): want %#v, got %#v", i, tt.expected, usr)
		}
	}
}
