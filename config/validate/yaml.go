package validate

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"

	"github.com/coreos/coreos-cloudinit/third_party/gopkg.in/yaml.v1"
)

var (
	YamlRules []rule = []rule{
		syntax,
		nodes,
	}
	yamlLineError = regexp.MustCompile(`^YAML error: line (?P<line>[[:digit:]]+): (?P<msg>.*)$`)
	yamlError     = regexp.MustCompile(`^YAML error: (?P<msg>.*)$`)
	yamlKey       = regexp.MustCompile(`^ *-? ?(?P<key>.*?):`)
)

func syntax(c context, v *validator) {
	if err := yaml.Unmarshal(c.content, &struct{}{}); err != nil {
		matches := yamlLineError.FindStringSubmatch(err.Error())
		if len(matches) > 0 {
			line, err := strconv.Atoi(matches[1])
			if err != nil {
				panic(err)
			}
			msg := matches[2]
			v.report.Error(c.line+line, msg)
			return
		}

		matches = yamlError.FindStringSubmatch(err.Error())
		if len(matches) > 0 {
			msg := matches[1]
			v.report.Error(c.line+1, msg)
			return
		}

		panic("couldn't parse yaml error")
	}
}

type node map[interface{}]interface{}

func nodes(c context, v *validator) {
	var n node
	if err := yaml.Unmarshal(c.content, &n); err == nil {
		checkNode(n, toNode(config.CloudConfig{}, ""), c, v)
	}
}

func toNode(s interface{}, prefix string) node {
	prefix += " "
	sv := reflect.ValueOf(s)

	if sv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("%T is not a struct (%s)", s, sv.Kind()))
	}

	n := make(node)
	for i := 0; i < sv.Type().NumField(); i++ {
		ft := sv.Type().Field(i)
		fv := sv.Field(i)
		k := ft.Tag.Get("yaml")

		if k == "-" || k == "" {
			continue
		}

		switch fv.Kind() {
		case reflect.Struct:
			n[k] = toNode(fv.Interface(), prefix)
		case reflect.Slice:
			et := ft.Type.Elem()

			switch et.Kind() {
			case reflect.Struct:
				n[k] = []node{toNode(reflect.New(et).Elem().Interface(), prefix)}
			default:
				n[k] = fv.Interface()
			}
		default:
			n[k] = fv.Interface()
		}
	}
	return n
}

func findKey(k string, c context) context {
	for len(c.content) > 0 {
		tokens := strings.SplitN(string(c.content), "\n", 2)
		line := tokens[0]

		c.line++
		if len(tokens) > 1 {
			c.content = []byte(tokens[1])
		} else {
			c.content = []byte{}
		}

		matches := yamlKey.FindStringSubmatch(line)
		if len(matches) > 0 && matches[1] == k {
			return c
		}
	}
	panic(fmt.Sprintf("key %q not found in content", k))
}

func checkNode(n, g node, c context, v *validator) {
	for k, sn := range n {
		c := findKey(fmt.Sprint(k), c)

		// Is the key expected?
		sg, ok := g[k]
		if !ok {
			v.report.Warning(c.line, fmt.Sprintf("unrecognized key %q", k))
			continue
		}
		if sg == nil {
			panic(fmt.Sprintf("reference node %q is nil", k))
		}

		// Should the object at the key be a slice of structs?
		if sg, ok := sg.([]node); ok {
			if sn, ok := sn.([]interface{}); ok {
				ssg := sg[0]
				for _, ssn := range sn {
					if ssn, ok := ssn.(map[interface{}]interface{}); ok {
						checkNode(ssn, ssg, c, v)
					} else {
						v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want struct)", k))
						continue
					}
				}
				continue
			} else {
				v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want []struct)", k))
				continue
			}
		}

		// Should the object at the key be a struct?
		if sg, ok := sg.(node); ok {
			if sn, ok := sn.(map[interface{}]interface{}); ok {
				checkNode(sn, sg, c, v)
				continue
			} else {
				v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want struct)", k))
				continue
			}
		}

		// Is the object at the key the correct type?
		if !isSameType(reflect.ValueOf(sn), reflect.ValueOf(sg)) {
			v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sg))
		}
	}
}

func isSameType(n, g reflect.Value) bool {
	for n.Kind() == reflect.Interface {
		n = n.Elem()
	}

	switch g.Kind() {
	case reflect.Bool:
		return n.Kind() == reflect.Bool
	case reflect.String:
		switch n.Kind() {
		case reflect.Bool, reflect.Int, reflect.String:
			return true
		default:
			return false
		}
	case reflect.Slice:
		if n.Kind() != g.Kind() {
			return false
		}

		sg := reflect.Indirect(reflect.New(g.Type().Elem()))
		for i := 0; i < n.Len(); i++ {
			if !isSameType(n.Index(i), sg) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("unhandled kind %s", g.Kind()))
	}
}
