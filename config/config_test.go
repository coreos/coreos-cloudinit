package config

import (
	"errors"
	"reflect"
	"testing"
)

func TestIsZero(t *testing.T) {
	for _, tt := range []struct {
		c     interface{}
		empty bool
	}{
		{struct{}{}, true},
		{struct{ a, b string }{}, true},
		{struct{ A, b string }{}, true},
		{struct{ A, B string }{}, true},
		{struct{ A string }{A: "hello"}, false},
		{struct{ A int }{}, true},
		{struct{ A int }{A: 1}, false},
	} {
		if empty := IsZero(tt.c); tt.empty != empty {
			t.Errorf("bad result (%q): want %q, got %q", tt.c, tt.empty, empty)
		}
	}
}

func TestAssertValid(t *testing.T) {
	for _, tt := range []struct {
		c   interface{}
		err error
	}{
		{struct{}{}, nil},
		{struct {
			A, b string `valid:"1,2"`
		}{}, nil},
		{struct {
			A, b string `valid:"1,2"`
		}{A: "1", b: "2"}, nil},
		{struct {
			A, b string `valid:"1,2"`
		}{A: "1", b: "hello"}, nil},
		{struct {
			A, b string `valid:"1,2"`
		}{A: "hello", b: "2"}, errors.New("invalid value \"hello\" for option \"A\" (valid options: \"1,2\")")},
		{struct {
			A, b int `valid:"1,2"`
		}{}, nil},
		{struct {
			A, b int `valid:"1,2"`
		}{A: 1, b: 2}, nil},
		{struct {
			A, b int `valid:"1,2"`
		}{A: 1, b: 9}, nil},
		{struct {
			A, b int `valid:"1,2"`
		}{A: 9, b: 2}, errors.New("invalid value \"9\" for option \"A\" (valid options: \"1,2\")")},
	} {
		if err := AssertValid(tt.c); !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad result (%q): want %q, got %q", tt.c, tt.err, err)
		}
	}
}
