package datasource

import (
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"
)

const maxTimeout = time.Second * 5

type Datasource interface {
	Fetch() ([]byte, error)
	Type() string
}

func fetchURL(url string) ([]byte, error) {
	resp, err := getWithExponentialBackoff(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBytes, nil
}

// getWithExponentialBackoff issues a GET to the specified URL. If the
// response is a non-2xx or produces an error, retry the GET forever using
// an exponential backoff.
func getWithExponentialBackoff(url string) (*http.Response, error) {
	var err error
	var resp *http.Response
	for i := 0; ; i++ {
		resp, err = http.Get(url)
		if err == nil && resp.StatusCode/100 == 2 {
			return resp, nil
		}
		duration := time.Millisecond * time.Duration((math.Pow(float64(2), float64(i)) * 100))
		if duration > maxTimeout {
			duration = maxTimeout
		}

		log.Printf("unable to fetch user-data from %s, try again in %s", url, duration)
		time.Sleep(duration)
	}
	return resp, err
}
