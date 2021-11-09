package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	*http.Client
	isInitialized bool
}

var pool = sync.Pool{
	New: func() interface{} {
		return &Client{&http.Client{}, false}
	},
}

func GetClient(timeout time.Duration) *Client {
	c := pool.Get().(*Client)
	if !c.isInitialized {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = 100
		t.MaxConnsPerHost = 100
		t.MaxIdleConnsPerHost = 100
		t.TLSClientConfig.InsecureSkipVerify = true
		c.Transport = t
		c.isInitialized = true
	}
	c.Timeout = timeout
	return c
}

func ReleaseClient(client *Client) {
	pool.Put(client)
}

func Request(target Target, timeout time.Duration) (int, int, error) {
	c := GetClient(timeout)
	defer ReleaseClient(c)

	req, err := http.NewRequest(target.Method, target.Host, nil)
	if err != nil {
		return 0, 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0")
	for _, h := range config.Headers {
		hParts := strings.Split(h, ":")
		if len(hParts) > 0 {
			req.Header.Set(hParts[0], strings.Join(hParts[1:],":"))
		}
	}

	var resp *http.Response
	resp, err = c.Do(req)
	if err != nil {
		return 0, 0, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	length := 0
	if err == nil {
		length = len(b)
	}
	return resp.StatusCode, length, nil
}
