package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	_ "embed"

	"jamel/internal/admin"
	"jamel/internal/admin/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

//go:embed server.crt
var TLSServerCert string

//go:embed admin.key
var TLSAdminKey string

//go:embed admin.crt
var TLSAdminCert string

func main() {
	creds, err := loadTLSCreds()
	if err != nil {
		fmt.Printf("failed to load server cert: %v\n", err)
	}
	_config := config.Must()
	conn, err := grpc.NewClient(
		_config.Server,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		fmt.Printf("error connect to server: %v\n", err)
	}
	defer conn.Close()
	_admin := admin.Must(
		_config.Username,
		_config.Password,
		conn,
	)
	if err := _admin.NewTaskFromFile(os.Args[1]); err != nil {
		log.Fatalln(err)
	}
}

func loadTLSCreds() (credentials.TransportCredentials, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(TLSServerCert)) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	clientCert, err := tls.X509KeyPair([]byte(TLSAdminCert), []byte(TLSAdminKey))
	if err != nil {
		return nil, err
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}
	return credentials.NewTLS(config), nil
}
