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

package config

import (
	"reflect"
	"testing"
)

func TestCommandValid(t *testing.T) {
	tests := []struct {
		value string

		isValid bool
	}{
		{value: "start", isValid: true},
		{value: "stop", isValid: true},
		{value: "restart", isValid: true},
		{value: "reload", isValid: true},
		{value: "try-restart", isValid: true},
		{value: "reload-or-restart", isValid: true},
		{value: "reload-or-try-restart", isValid: true},
		{value: "tryrestart", isValid: false},
		{value: "unknown", isValid: false},
	}

	for _, tt := range tests {
		isValid := (nil == AssertStructValid(Unit{Command: tt.value}))
		if tt.isValid != isValid {
			t.Errorf("bad assert (%s): want %t, got %t", tt.value, tt.isValid, isValid)
		}
	}
}

func TestUnitMerge(t *testing.T) {
	tests := []struct {
		inputs   []Unit
		expected Unit
	}{
		{
			inputs: []Unit{
				Unit{DropIns: []UnitDropIn{UnitDropIn{Name: "unit1"}}},
				Unit{DropIns: []UnitDropIn{UnitDropIn{Name: "unit2"}}},
				Unit{DropIns: []UnitDropIn{UnitDropIn{Name: "unit1"}}},
			},
			expected: Unit{DropIns: []UnitDropIn{UnitDropIn{Name: "unit1"}, UnitDropIn{Name: "unit2"}}},
		},
	}
	for i, tt := range tests {
		u := tt.inputs[0]
		for _, t := range tt.inputs[1:] {
			u.Merge(t)
		}
		if !reflect.DeepEqual(tt.expected, u) {
			t.Errorf("bad unit (test case #%d): want %#v, got %#v", i, tt.expected, u)
		}
	}
}
