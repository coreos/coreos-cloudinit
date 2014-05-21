package datasource

import (
	"testing"
)

func TestParseCmdlineCloudConfigFound(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			"cloud-config-url=example.com",
			"example.com",
		},
		{
			"cloud_config_url=example.com",
			"example.com",
		},
		{
			"cloud-config-url cloud-config-url=example.com",
			"example.com",
		},
		{
			"cloud-config-url= cloud-config-url=example.com",
			"example.com",
		},
		{
			"cloud-config-url=one.example.com cloud-config-url=two.example.com",
			"two.example.com",
		},
		{
			"foo=bar cloud-config-url=example.com ping=pong",
			"example.com",
		},
	}

	for i, tt := range tests {
		output, err := findCloudConfigURL(tt.input)
		if output != tt.expect {
			t.Errorf("Test case %d failed: %s != %s", i, output, tt.expect)
		}
		if err != nil {
			t.Errorf("Test case %d produced error: %v", i, err)
		}
	}
}
