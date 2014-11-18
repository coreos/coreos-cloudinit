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

package validate

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"

	"github.com/coreos/coreos-cloudinit/Godeps/_workspace/src/gopkg.in/yaml.v1"
)

var (
	yamlLineError = regexp.MustCompile(`^YAML error: line (?P<line>[[:digit:]]+): (?P<msg>.*)$`)
	yamlError     = regexp.MustCompile(`^YAML error: (?P<msg>.*)$`)
)

// Validate runs a series of validation tests against the given userdata and
// returns a report detailing all of the issues. Presently, only cloud-configs
// can be validated.
func Validate(userdataBytes []byte) (Report, error) {
	switch {
	case config.IsScript(string(userdataBytes)):
		return Report{}, nil
	case config.IsCloudConfig(string(userdataBytes)):
		return validateCloudConfig(userdataBytes, Rules)
	default:
		return Report{entries: []Entry{
			Entry{kind: entryError, message: `must be "#cloud-config" or begin with "#!"`, line: 1},
		}}, nil
	}
}

// validateCloudConfig runs all of the validation rules in Rules and returns
// the resulting report and any errors encountered.
func validateCloudConfig(config []byte, rules []rule) (report Report, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	c, err := parseCloudConfig(config, &report)
	if err != nil {
		return report, err
	}

	c = normalizeNodeNames(c, &report)
	for _, r := range rules {
		r(c, &report)
	}
	return report, nil
}

// parseCloudConfig parses the provided config into a node structure and logs
// any parsing issues into the provided report. Unrecoverable errors are
// returned as an error.
func parseCloudConfig(config []byte, report *Report) (n node, err error) {
	var raw map[interface{}]interface{}
	if err := yaml.Unmarshal(config, &raw); err != nil {
		matches := yamlLineError.FindStringSubmatch(err.Error())
		if len(matches) == 3 {
			line, err := strconv.Atoi(matches[1])
			if err != nil {
				return n, err
			}
			msg := matches[2]
			report.Error(line, msg)
			return n, nil
		}

		matches = yamlError.FindStringSubmatch(err.Error())
		if len(matches) == 2 {
			report.Error(1, matches[1])
			return n, nil
		}

		return n, errors.New("couldn't parse yaml error")
	}

	return NewNode(raw, NewContext(config)), nil
}

// normalizeNodeNames replaces all occurences of '-' with '_' within key names
// and makes a note of each replacement in the report.
func normalizeNodeNames(node node, report *Report) node {
	if strings.Contains(node.name, "-") {
		report.Info(node.line, fmt.Sprintf("%q uses '-' instead of '_'", node.name))
		node.name = strings.Replace(node.name, "-", "_", -1)
	}
	for i := range node.children {
		node.children[i] = normalizeNodeNames(node.children[i], report)
	}
	return node
}
