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

type Unit struct {
	Name    string       `yaml:"name,omitempty"`
	Mask    bool         `yaml:"mask,omitempty"`
	Enable  bool         `yaml:"enable,omitempty"`
	Runtime bool         `yaml:"runtime,omitempty"`
	Content string       `yaml:"content,omitempty"`
	Command string       `yaml:"command,omitempty" valid:"^(start|stop|restart|reload|try-restart|reload-or-restart|reload-or-try-restart)$"`
	DropIns []UnitDropIn `yaml:"drop_ins,omitempty"`
}

type UnitDropIn struct {
	Name    string `yaml:"name,omitempty"`
	Content string `yaml:"content,omitempty"`
}
