package validate

import (
	"reflect"
	"testing"
)

func TestSyntax(t *testing.T) {
	for _, tt := range []struct {
		c string
		e []Entry
	}{
		{},
		{
			c: "	",
			e: []Entry{{entryError, "found character that cannot start any token", 1}},
		},
		{
			c: "a:\na",
			e: []Entry{{entryError, "could not find expected ':'", 2}},
		},
	} {
		v := validator{report: &Report{}}
		syntax(context{content: []byte(tt.c)}, &v)

		if e := v.report.Entries(); !reflect.DeepEqual(tt.e, e) {
			t.Fatalf("bad report (%s): want %#v, got %#v", tt.c, tt.e, e)
		}
	}
}

func TestNodes(t *testing.T) {
	for i, tt := range []struct {
		c string
		e []Entry
	}{
		{},

		// Test for unrecognized keys
		{
			c: "test:",
			e: []Entry{{entryWarning, "unrecognized key \"test\"", 1}},
		},
		{
			c: "coreos:\n  etcd:\n    bad:",
			e: []Entry{{entryWarning, "unrecognized key \"bad\"", 3}},
		},
		{
			c: "coreos:\n  etcd:\n    discovery: good",
		},

		// Test for incorrect types
		// Want boolean
		{
			c: "coreos:\n  units:\n    - enable: true",
		},
		{
			c: "coreos:\n  units:\n    - enable: 4",
			e: []Entry{{entryWarning, "incorrect type for \"enable\" (want bool)", 3}},
		},
		{
			c: "coreos:\n  units:\n    - enable: bad",
			e: []Entry{{entryWarning, "incorrect type for \"enable\" (want bool)", 3}},
		},
		{
			c: "coreos:\n  units:\n    - enable:\n        bad:",
			e: []Entry{{entryWarning, "incorrect type for \"enable\" (want bool)", 3}},
		},
		{
			c: "coreos:\n  units:\n    - enable:\n      - bad",
			e: []Entry{{entryWarning, "incorrect type for \"enable\" (want bool)", 3}},
		},
		// Want string
		{
			c: "hostname: true",
		},
		{
			c: "hostname: 4",
		},
		{
			c: "hostname: host",
		},
		{
			c: "hostname:\n  name:",
			e: []Entry{{entryWarning, "incorrect type for \"hostname\" (want string)", 1}},
		},
		{
			c: "hostname:\n  - name",
			e: []Entry{{entryWarning, "incorrect type for \"hostname\" (want string)", 1}},
		},
		// Want struct
		{
			c: "coreos: true",
			e: []Entry{{entryWarning, "incorrect type for \"coreos\" (want struct)", 1}},
		},
		{
			c: "coreos: 4",
			e: []Entry{{entryWarning, "incorrect type for \"coreos\" (want struct)", 1}},
		},
		{
			c: "coreos: hello",
			e: []Entry{{entryWarning, "incorrect type for \"coreos\" (want struct)", 1}},
		},
		{
			c: "coreos:\n  etcd:\n    discovery: fire in the disco",
		},
		{
			c: "coreos:\n  - hello",
			e: []Entry{{entryWarning, "incorrect type for \"coreos\" (want struct)", 1}},
		},
		// Want []string
		{
			c: "ssh_authorized_keys: true",
			e: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		{
			c: "ssh_authorized_keys: 4",
			e: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		{
			c: "ssh_authorized_keys: key",
			e: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		{
			c: "ssh_authorized_keys:\n  key: value",
			e: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		{
			c: "ssh_authorized_keys:\n  - key",
		},
		{
			c: "ssh_authorized_keys:\n  - key: value",
			e: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		// Want []struct
		{
			c: "users:\n  true",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want []struct)", 1}},
		},
		{
			c: "users:\n  4",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want []struct)", 1}},
		},
		{
			c: "users:\n  bad",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want []struct)", 1}},
		},
		{
			c: "users:\n  bad:",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want []struct)", 1}},
		},
		{
			c: "users:\n  - name: good",
		},
		// Want struct within array
		{
			c: "users:\n  - true",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want struct)", 1}},
		},
		{
			c: "users:\n  - 4",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want struct)", 1}},
		},
		{
			c: "users:\n  - bad",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want struct)", 1}},
		},
		{
			c: "users:\n  - - bad",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want struct)", 1}},
		},
	} {
		v := validator{report: &Report{}}
		nodes(context{content: []byte(tt.c)}, &v)

		if e := v.report.Entries(); !reflect.DeepEqual(tt.e, e) {
			t.Fatalf("bad report (%d, %s): want %#v, got %#v", i, tt.c, tt.e, e)
		}
	}
}
