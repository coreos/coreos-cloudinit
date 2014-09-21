package validate

import (
	"bytes"
	"reflect"
	"testing"
)

func TestEntry(t *testing.T) {
	for _, tt := range []struct {
		e Entry
		s string
		j []byte
	}{
		{
			Entry{entryWarning, "test warning", 1},
			"line 1: warning: test warning",
			[]byte(`{"kind":"warning","line":1,"message":"test warning"}`),
		},
		{
			Entry{entryError, "test error", 2},
			"line 2: error: test error",
			[]byte(`{"kind":"error","line":2,"message":"test error"}`),
		},
	} {
		if s := tt.e.String(); tt.s != s {
			t.Errorf("bad string (%q): want %q, got %q", tt.e, tt.s, s)
		}
		j, err := tt.e.MarshalJSON()
		if err != nil {
			t.Errorf("bad error (%q): want %q, got %q", tt.e, nil, err)
		}
		if !bytes.Equal(tt.j, j) {
			t.Errorf("bad JSON (%q): want %q, got %q", tt.e, tt.j, j)
		}
	}
}

func TestReport(t *testing.T) {
	type reportFunc struct {
		fn      func(*Report, int, string)
		line    int
		message string
	}
	for _, tt := range []struct {
		fs []reportFunc
		es []Entry
	}{
		{
			[]reportFunc{
				{(*Report).Warning, 1, "test warning 1"},
				{(*Report).Error, 2, "test error 2"},
				{(*Report).Warning, 10, "test warning 10"},
			},
			[]Entry{
				Entry{entryWarning, "test warning 1", 1},
				Entry{entryError, "test error 2", 2},
				Entry{entryWarning, "test warning 10", 10},
			},
		},
	} {
		r := Report{}
		for _, f := range tt.fs {
			f.fn(&r, f.line, f.message)
		}
		if es := r.Entries(); !reflect.DeepEqual(tt.es, es) {
			t.Errorf("bad entries (%q): want %#v, got %#v", tt.fs, tt.es, es)
		}
	}
}
