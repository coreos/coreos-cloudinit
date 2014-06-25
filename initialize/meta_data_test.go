package initialize

import "reflect"
import "testing"

func TestExtractIPsFromMetadata(t *testing.T) {
	for i, tt := range []struct {
		in  []byte
		err bool
		out map[string]string
	}{
		{
			[]byte(`{"public-ipv4": "12.34.56.78", "local-ipv4": "1.2.3.4"}`),
			false,
			map[string]string{"$public_ipv4": "12.34.56.78", "$private_ipv4": "1.2.3.4"},
		},
		{
			[]byte(`{"local-ipv4": "127.0.0.1", "something_else": "don't care"}`),
			false,
			map[string]string{"$private_ipv4": "127.0.0.1"},
		},
		{
			[]byte(`garbage`),
			true,
			nil,
		},
	} {
		got, err := ExtractIPsFromMetadata(tt.in)
		if (err != nil) != tt.err {
			t.Errorf("bad error state (got %t, want %t)", err != nil, tt.err)
		}
		if !reflect.DeepEqual(got, tt.out) {
			t.Errorf("case %d: got %s, want %s", i, got, tt.out)
		}
	}
}
