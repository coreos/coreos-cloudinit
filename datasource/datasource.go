package datasource

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	neturl "net/url"
	"strings"
	"time"
)

const (
	HTTP_2xx = 2
	HTTP_4xx = 4

	maxTimeout = time.Second * 5
	maxRetries = 15
)

type Datasource interface {
	Fetch() ([]byte, error)
	Type() string
}

// HTTP client timeout
// This one is low since exponential backoff will kick off too.
var timeout = time.Duration(2) * time.Second

func dialTimeout(network, addr string) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	c, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	c.SetDeadline(deadline)
	return c, nil
}

// Fetches user-data url with support for exponential backoff and maximum retries
func fetchURL(rawurl string) ([]byte, error) {
	if rawurl == "" {
		return nil, errors.New("user-data URL is empty. Skipping.")
	}

	url, err := neturl.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	// Unfortunately, url.Parse is too generic to throw errors if a URL does not
	// have a valid HTTP scheme. So, we have to do this extra validation
	if !strings.HasPrefix(url.Scheme, "http") {
		return nil, fmt.Errorf("user-data URL %s does not have a valid HTTP scheme. Skipping.", rawurl)
	}

	userdataURL := url.String()

	// We need to create our own client in order to add timeout support.
	// TODO(c4milo) Replace it once Go 1.3 is officially used by CoreOS
	// More info: https://code.google.com/p/go/source/detail?r=ada6f2d5f99f
	transport := &http.Transport{
		Dial: dialTimeout,
	}

	client := &http.Client{
		Transport: transport,
	}

	for retry := 0; retry <= maxRetries; retry++ {
		log.Printf("Fetching user-data from %s. Attempt #%d", userdataURL, retry)

		resp, err := client.Get(userdataURL)

		if err == nil {
			defer resp.Body.Close()
			status := resp.StatusCode / 100

			if status == HTTP_2xx {
				return ioutil.ReadAll(resp.Body)
			}

			if status == HTTP_4xx {
				return nil, fmt.Errorf("user-data not found. HTTP status code: %d", resp.StatusCode)
			}

			log.Printf("user-data not found. HTTP status code: %d", resp.StatusCode)
		} else {
			log.Printf("unable to fetch user-data: %s", err.Error())
		}

		duration := time.Millisecond * time.Duration((math.Pow(float64(2), float64(retry)) * 100))
		if duration > maxTimeout {
			duration = maxTimeout
		}

		time.Sleep(duration)
	}

	return nil, fmt.Errorf("unable to fetch user-data. Maximum retries reached: %d", maxRetries)
}
