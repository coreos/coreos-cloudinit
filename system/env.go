package system

import (
	"fmt"
	"reflect"
)

// dropinContents generates the contents for a drop-in unit given the config.
// The argument must be a struct from the 'config' package.
func dropinContents(e interface{}) string {
	et := reflect.TypeOf(e)
	ev := reflect.ValueOf(e)

	var out string
	for i := 0; i < et.NumField(); i++ {
		if val := ev.Field(i).String(); val != "" {
			key := et.Field(i).Tag.Get("env")
			out += fmt.Sprintf("Environment=\"%s=%s\"\n", key, val)
		}
	}

	if out == "" {
		return ""
	}
	return "[Service]\n" + out
}
