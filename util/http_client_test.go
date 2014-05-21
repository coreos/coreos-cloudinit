package util

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var expBackoffTests = []struct {
	count int
	body  string
}{
	{0, "number of attempts: 0"},
	{1, "number of attempts: 1"},
	{2, "number of attempts: 2"},
}

// Test exponential backoff and that it continues retrying if a 5xx response is
// received
func TestFetchURLExpBackOff(t *testing.T) {
	client := NewHttpClient()

	for i, tt := range expBackoffTests {
		mux := http.NewServeMux()
		count := 0
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if count == tt.count {
				io.WriteString(w, fmt.Sprintf("number of attempts: %d", count))
				return
			}
			count++
			http.Error(w, "", 500)
		})
		ts := httptest.NewServer(mux)
		defer ts.Close()

		data, err := client.Get(ts.URL)
		if err != nil {
			t.Errorf("Test case %d produced error: %v", i, err)
		}

		if count != tt.count {
			t.Errorf("Test case %d failed: %d != %d", i, count, tt.count)
		}

		if string(data) != tt.body {
			t.Errorf("Test case %d failed: %s != %s", i, tt.body, data)
		}
	}
}

// Test that it stops retrying if a 4xx response comes back
func TestFetchURL4xx(t *testing.T) {
	client := NewHttpClient()
	retries := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retries++
		http.Error(w, "", 404)
	}))
	defer ts.Close()

	_, err := client.Get(ts.URL)
	if err == nil {
		t.Errorf("Incorrect result\ngot:  %s\nwant: %s", err.Error(), "Not found. HTTP status code: 404")
	}

	if retries > 1 {
		t.Errorf("Number of retries:\n%d\nExpected number of retries:\n%s", retries, 1)
	}
}

// Test that it fetches and returns user-data just fine
func TestFetchURL2xx(t *testing.T) {
	var cloudcfg = `
#cloud-config
coreos: 
	oem:
	    id: test
	    name: CoreOS.box for Test
	    version-id: %VERSION_ID%+%BUILD_ID%
	    home-url: https://github.com/coreos/coreos-cloudinit
	    bug-report-url: https://github.com/coreos/coreos-cloudinit
	update:
		reboot-strategy: best-effort
`

	client := NewHttpClient()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, cloudcfg)
	}))
	defer ts.Close()

	data, err := client.Get(ts.URL)
	if err != nil {
		t.Errorf("Incorrect result\ngot:  %v\nwant: %v", err, nil)
	}

	if string(data) != cloudcfg {
		t.Errorf("Incorrect result\ngot:  %s\nwant: %s", string(data), cloudcfg)
	}
}

// Test attempt to fetching using malformed URL
func TestFetchURLMalformed(t *testing.T) {
	client := NewHttpClient()

	var tests = []struct {
		url  string
		want string
	}{
		{"boo", "URL boo does not have a valid HTTP scheme. Skipping."},
		{"mailto://boo", "URL mailto://boo does not have a valid HTTP scheme. Skipping."},
		{"ftp://boo", "URL ftp://boo does not have a valid HTTP scheme. Skipping."},
		{"", "URL is empty. Skipping."},
	}

	for _, test := range tests {
		_, err := client.Get(test.url)
		if err == nil || err.Error() != test.want {
			t.Errorf("Incorrect result\ngot:  %v\nwant: %v", err, test.want)
		}
	}
}
