package validate

import (
	"errors"
	"reflect"
	"testing"
)

func TestRunTest(t *testing.T) {
	for _, tt := range []struct {
		t   test
		err error
	}{
		{test{context{}, func(_ context, _ *validator) {}}, nil},
		{test{context{}, func(_ context, _ *validator) { panic("hi") }}, errors.New("hi")},
	} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("bad panic (%+v)", tt.t)
				}
			}()
			if err := runTest(tt.t, &validator{}); !reflect.DeepEqual(tt.err, err) {
				t.Errorf("bad error (%+v): want %q, got %q", tt.t, tt.err, err)
			}
		}()
	}
}

func TestBaseRule(t *testing.T) {
	for _, tt := range []struct {
		c  string
		es []Entry
	}{
		{"", []Entry{Entry{entryError, `must be "#cloud-config" or "#!"`, 1}}},
		{"hello", []Entry{Entry{entryError, `must be "#cloud-config" or "#!"`, 1}}},
		{"#cloud-config", nil},
		{"#cloud-config hello", []Entry{Entry{entryError, `must be "#cloud-config" or "#!"`, 1}}},
		{" #cloud-config", []Entry{Entry{entryError, `must be "#cloud-config" or "#!"`, 1}}},
		{"#!", nil},
		{" #!", []Entry{Entry{entryError, `must be "#cloud-config" or "#!"`, 1}}},
		{"#! /bin/hello", nil},
	} {
		v := validator{
			report: &Report{},
		}
		baseRule(context{content: []byte(tt.c)}, &v)
		if es := v.report.Entries(); !reflect.DeepEqual(tt.es, es) {
			t.Errorf("bad entries (%q): want %#v, got %#v", tt.c, tt.es, es)
		}
	}
}
