package datasource

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestProcCmdlineAndFetchConfig(t *testing.T) {

	var (
		ProcCmdlineTmpl    = "foo=bar cloud-config-url=%s/config\n"
		CloudConfigContent = "#cloud-config\n"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.RequestURI == "/config" {
			fmt.Fprint(w, CloudConfigContent)
		}
	}))
	defer ts.Close()

	file, err := ioutil.TempFile(os.TempDir(), "test_proc_cmdline")
	defer os.Remove(file.Name())
	if err != nil {
		t.Errorf("Test produced error: %v", err)
	}
	_, err = file.Write([]byte(fmt.Sprintf(ProcCmdlineTmpl, ts.URL)))
	if err != nil {
		t.Errorf("Test produced error: %v", err)
	}

	p := NewProcCmdline()
	p.Location = file.Name()
	cfg, err := p.FetchUserdata()
	if err != nil {
		t.Errorf("Test produced error: %v", err)
	}

	if string(cfg) != CloudConfigContent {
		t.Errorf("Test failed, response body: %s != %s", cfg, CloudConfigContent)
	}
}
