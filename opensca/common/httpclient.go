package common

import (
	"crypto/tls"
	"net/http"
	"time"
)

var HttpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        50,
		MaxConnsPerHost:     50,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	},
	Timeout: 60 * time.Second,
}

func InitHttpClient(insecure bool) {
	HttpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = insecure
}
