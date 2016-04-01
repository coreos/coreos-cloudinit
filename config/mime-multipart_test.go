package config

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func sameMimeMultiParts(a, b MimeMultiPart) bool {
	if len(a.Scripts) != len(b.Scripts) {
		return false
	}
	for i, _ := range a.Scripts {
		if !bytes.Equal(*a.Scripts[i], *b.Scripts[i]) {
			return false
		}
	}
	if !reflect.DeepEqual(a.Config, b.Config) {
		return false
	}
	return true
}

func TestNewMimeMultiPart(t *testing.T) {
	testScript := `#! /bin/bash
cat <<'EOF' >/var/lib/rightscale-identity
account='67972'
api_hostname='us-3.rightscale.com'
client_id='18643426003'
client_secret='a13bbcc595f773459fd9664da52480d2057d43e1'
EOF
chmod 0600 /var/lib/rightscale-identity`
	testConfig := "#cloud-config\nwrite_files:\n  - permissions: 0744"
	t1, _ := NewScript(testScript)
	t2, _ := NewCloudConfig(testConfig)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(testScript))
	}))

	tests := []struct {
		contents string
		mmp      MimeMultiPart
	}{
		{
			contents: fmt.Sprintf(`Content-ID: <a4cdaxbfbd03yc84am2be4@ek4znmbfbd03yc84am2bfw.local>
Content-Type: multipart/mixed; boundary=Boundary_ayn1dgbfbd03yc84am2azq
MIME-Version: 1.0

--Boundary_ayn1dgbfbd03yc84am2azq
Content-ID: <98w9o8bfbd03yc84am2cpg@a17vgbfbd03yc84am2cra.local>
Content-Type: text/x-shellscript

%s
--Boundary_ayn1dgbfbd03yc84am2azq
Content-ID: <c5vsnxbfbd03yc84am2cvy@cxx0y3bfbd03yc84am2cx6.local>
Content-Type: text/cloud-config

%s
--Boundary_ayn1dgbfbd03yc84am2azq--
`, testScript, testConfig),
			mmp: MimeMultiPart{Scripts: []*Script{t1}, Config: t2},
		},
		{
			contents: fmt.Sprintf(`Content-ID: <a4cdaxbfbd03yc84am2be4@ek4znmbfbd03yc84am2bfw.local>
Content-Type: multipart/mixed; boundary=Boundary_ayn1dgbfbd03yc84am2azq
MIME-Version: 1.0

--Boundary_ayn1dgbfbd03yc84am2azq
Content-ID: <98w9o8bfbd03yc84am2cpg@a17vgbfbd03yc84am2cra.local>
Content-Type: text/x-include-url

%s
--Boundary_ayn1dgbfbd03yc84am2azq
Content-ID: <c5vsnxbfbd03yc84am2cvy@cxx0y3bfbd03yc84am2cx6.local>
Content-Type: text/cloud-config

%s
--Boundary_ayn1dgbfbd03yc84am2azq--
`, testServer.URL, testConfig),
			mmp: MimeMultiPart{Scripts: []*Script{t1}, Config: t2},
		},
	}

	for i, tt := range tests {
		mmp, err := NewMimeMultiPart(tt.contents)
		if err != nil {
			t.Errorf("bad error (test case #%d): want %v, got %s", i, nil, err)
		}
		if !sameMimeMultiParts(tt.mmp, *mmp) {
			t.Errorf("bad mime-multipart (test case #%d): want %#v, got %#v", i, tt.mmp, mmp)
		}
	}
}
