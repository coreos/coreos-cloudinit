package validate

import (
	"fmt"
	"strings"
)

type Reporter interface {
	Error(line int, message string)
	Warning(line int, message string)
	Entries() []Entry
}

type context struct {
	content []byte
	line    int
}

type rule func(context context, validator *validator)

type test struct {
	context context
	rule    rule
}

type validator struct {
	report Reporter
	tests  []test
}

func (v *validator) addRules(c context, rs ...rule) {
	for _, r := range rs {
		v.tests = append(v.tests, test{c, r})
	}
}

func Validate(config []byte) (Reporter, error) {
	v := &validator{&Report{}, []test{{context{content: config}, baseRule}}}

	for len(v.tests) > 0 {
		t := v.tests[0]
		v.tests = v.tests[1:]

		if err := runTest(t, v); err != nil {
			return v.report, err
		}
	}

	return v.report, nil
}

func runTest(t test, v *validator) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	t.rule(t.context, v)
	return nil
}

func baseRule(c context, v *validator) {
	header := strings.SplitN(string(c.content), "\n", 2)[0]
	if header == "#cloud-config" {
		v.addRules(c, YamlRules...)
	} else if !strings.HasPrefix(header, "#!") {
		v.report.Error(c.line+1, `must be "#cloud-config" or "#!"`)
	}
}
