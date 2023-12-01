package common

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/cmd/config"
)

var HttpClient = http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        50,
		MaxConnsPerHost:     50,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.Conf().Optional.Insecure,
		},
	},
	Timeout: 60 * time.Second,
}
