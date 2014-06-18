package initialize

import (
	"os"
	"testing"
)

func TestEnvironmentApply(t *testing.T) {
	os.Setenv("COREOS_PUBLIC_IPV4", "192.0.2.3")
	os.Setenv("COREOS_PRIVATE_IPV4", "192.0.2.203")
	env := NewEnvironment("./", "./", "./", "", "")
	input := `[Service]
ExecStart=/usr/bin/echo "$public_ipv4"
ExecStop=/usr/bin/echo $private_ipv4
ExecStop=/usr/bin/echo $unknown
`
	expected := `[Service]
ExecStart=/usr/bin/echo "192.0.2.3"
ExecStop=/usr/bin/echo 192.0.2.203
ExecStop=/usr/bin/echo $unknown
`

	output := env.Apply(input)
	if output != expected {
		t.Fatalf("Environment incorrectly applied.\nOutput:\n%s\nExpected:\n%s", output, expected)
	}
}
