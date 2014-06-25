package initialize

import (
	"os"
	"testing"
)

func TestEnvironmentApply(t *testing.T) {
	os.Setenv("COREOS_PUBLIC_IPV4", "1.2.3.4")
	os.Setenv("COREOS_PRIVATE_IPV4", "5.6.7.8")
	for _, tt := range []struct {
		subs  map[string]string
		input string
		out   string
	}{
		{
			// Substituting both values directly should always take precedence
			// over environment variables
			map[string]string{
				"$public_ipv4":  "192.0.2.3",
				"$private_ipv4": "192.0.2.203",
			},
			`[Service]
ExecStart=/usr/bin/echo "$public_ipv4"
ExecStop=/usr/bin/echo $private_ipv4
ExecStop=/usr/bin/echo $unknown`,
			`[Service]
ExecStart=/usr/bin/echo "192.0.2.3"
ExecStop=/usr/bin/echo 192.0.2.203
ExecStop=/usr/bin/echo $unknown`,
		},
		{
			// Substituting one value directly while falling back with the other
			map[string]string{"$private_ipv4": "127.0.0.1"},
			"$private_ipv4\n$public_ipv4",
			"127.0.0.1\n1.2.3.4",
		},
		{
			// Falling back to environment variables for both values
			map[string]string{"foo": "bar"},
			"$private_ipv4\n$public_ipv4",
			"5.6.7.8\n1.2.3.4",
		},
		{
			// No substitutions
			nil,
			"$private_ipv4\nfoobar",
			"5.6.7.8\nfoobar",
		},
	} {

		env := NewEnvironment("./", "./", "./", "", "", tt.subs)
		got := env.Apply(tt.input)
		if got != tt.out {
			t.Fatalf("Environment incorrectly applied.\ngot:\n%s\nwant:\n%s", got, tt.out)
		}
	}
}
