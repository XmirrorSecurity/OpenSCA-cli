package common

import (
	"crypto/tls"
	"net/http"
	"time"
)

var HttpClient = http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        50,
		MaxConnsPerHost:     50,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
	Timeout: 10 * time.Second,
}
