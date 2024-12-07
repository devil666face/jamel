package http

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"jamel/pkg/fs"
	"net/http"
)

func InsecureClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func SaveCAs(urls []string, file string) error {
	var certs []byte

	for _, url := range urls {
		cert, err := GetCA(url)
		if err != nil {
			return fmt.Errorf("failed to get cert: %w", err)
		}
		certs = append(certs, cert...)
	}

	if err := fs.WriteFile(file, certs); err != nil {
		return fmt.Errorf("failed to prepare cert for update: %w", err)
	}
	return nil
}

func GetCA(url string) ([]byte, error) {
	var client = InsecureClient()

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	tlsstate := resp.TLS
	if tlsstate == nil {
		return nil, fmt.Errorf("no TLS connection state available")
	}

	if len(tlsstate.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no peer certificates found")
	}

	var (
		buf     bytes.Buffer
		certpem = &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: tlsstate.PeerCertificates[len(tlsstate.PeerCertificates)-1].Raw,
		}
	)
	if err := pem.Encode(&buf, certpem); err != nil {
		return nil, fmt.Errorf("failed to encode cert: %w", err)
	}

	return buf.Bytes(), nil
}
