package config

import (
	"fmt"
	"reflect"
	"strings"
)

// IsZero returns whether or not the parameter is the zero value for its type.
// If the parameter is a struct, only the exported fields are considered.
func IsZero(c interface{}) bool {
	return isZero(reflect.ValueOf(c))
}

// AssertValid checks the fields in the structure and makes sure that they
// contain valid values as specified by the 'valid' flag. Empty fields are
// implicitly valid.
func AssertValid(c interface{}) error {
	ct := reflect.TypeOf(c)
	cv := reflect.ValueOf(c)
	for i := 0; i < ct.NumField(); i++ {
		ft := ct.Field(i)
		if !isFieldExported(ft) {
			continue
		}

		valid := ft.Tag.Get("valid")
		val := cv.Field(i)
		if !isValid(val, valid) {
			return fmt.Errorf("invalid value \"%v\" for option %q (valid options: %q)", val.Interface(), ft.Name, valid)
		}
	}
	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Struct:
		vt := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if isFieldExported(vt.Field(i)) && !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return v.Interface() == reflect.Zero(v.Type()).Interface()
	}
}

func isFieldExported(f reflect.StructField) bool {
	return f.PkgPath == ""
}

func isValid(v reflect.Value, valid string) bool {
	if valid == "" || isZero(v) {
		return true
	}
	vs := fmt.Sprintf("%v", v.Interface())
	for _, valid := range strings.Split(valid, ",") {
		if vs == valid {
			return true
		}
	}
	return false
}
