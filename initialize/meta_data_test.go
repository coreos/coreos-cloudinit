package initialize

import "reflect"
import "testing"

func TestParseMetadata(t *testing.T) {
	for i, tt := range []struct {
		in   string
		want *CloudConfig
		err  bool
	}{
		{"", nil, false},
		{`garbage, invalid json`, nil, true},
		{`{"foo": "bar"}`, &CloudConfig{}, false},
		{`{"network_config": {"content_path": "asdf"}}`, &CloudConfig{NetworkConfigPath: "asdf"}, false},
		{`{"hostname": "turkleton"}`, &CloudConfig{Hostname: "turkleton"}, false},
		{`{"public_keys": {"jack": "jill", "bob": "alice"}}`, &CloudConfig{SSHAuthorizedKeys: []string{"jill", "alice"}}, false},
		{`{"unknown": "thing", "hostname": "my_host", "public_keys": {"do": "re", "mi": "fa"}, "network_config": {"content_path": "/root", "blah": "zzz"}}`, &CloudConfig{SSHAuthorizedKeys: []string{"re", "fa"}, Hostname: "my_host", NetworkConfigPath: "/root"}, false},
	} {
		got, err := ParseMetaData(tt.in)
		if tt.err != (err != nil) {
			t.Errorf("case #%d: bad error state: got %t, want %t (err=%v)", i, (err != nil), tt.err, err)
		}
		if got == nil {
			if tt.want != nil {
				t.Errorf("case #%d: unexpected nil output", i)
			}
		} else if tt.want == nil {
			t.Errorf("case #%d: unexpected non-nil output", i)
		} else {
			if !reflect.DeepEqual(*got, *tt.want) {
				t.Errorf("case #%d: bad output:\ngot\n%v\nwant\n%v", i, *got, *tt.want)
			}
		}
	}

}

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
