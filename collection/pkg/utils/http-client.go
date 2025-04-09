package utils

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

func CreateHTTPClient(proxyUrl *url.URL) *http.Client {

	var transport *http.Transport
	tlsConfig := &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // Example - Find correct values
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, // Example - Find correct values
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,       // Example - Find correct values
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,       // Example - Find correct values
		},
		CurvePreferences: []tls.CurveID{
			tls.CurveP256, // Example - Find correct values
			tls.CurveP384, // Example - Find correct values
		},
		MinVersion:         tls.VersionTLS12, // Example - Find correct values
		MaxVersion:         tls.VersionTLS13, // Example - Find correct values
		InsecureSkipVerify: true,
	}

	if proxyUrl != nil {
		transport = &http.Transport{
			Proxy:           http.ProxyURL(proxyUrl),
			TLSClientConfig: tlsConfig,
		}
	} else {
		transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
	return client
}
