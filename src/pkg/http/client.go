package http

import (
	"crypto/tls"
	"net/http"
	"time"
)

// SecureHTTPClient creates HTTP client with security settings
type SecureHTTPClient struct {
	client *http.Client
}

// NewSecureHTTPClient creates new secure HTTP client
func NewSecureHTTPClient(timeout time.Duration) *SecureHTTPClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			},
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	return &SecureHTTPClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

// Do executes HTTP request
func (c *SecureHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// GetClient returns underlying HTTP client
func (c *SecureHTTPClient) GetClient() *http.Client {
	return c.client
}
