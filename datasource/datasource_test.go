package datasource

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

func TestGetWithExponentialBackoff(t *testing.T) {
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
		data, err := fetchURL(ts.URL)
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
