package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	_ "embed"

	"jamel/internal/server"
	"jamel/internal/server/config"
	"jamel/internal/server/grpc/handler"
	"jamel/internal/server/service/store"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

//go:embed server.key
var TLSServerKey []byte

//go:embed server.crt
var TLSServerCert []byte

//go:embed client.crt
var TLSClientCert []byte

//go:embed admin.crt
var TLSAdminCert []byte

func main() {
	if err := os.MkdirAll(handler.StaticDir, 0o755); err != nil {
		log.Fatalf("failed to create dir for static files: %v", err)
	}

	creds, err := loadTLSCreds()
	if err != nil {
		log.Fatalf("failed to load tls cert or key: %v", err)
	}
	_config := config.Must()
	_store := store.Must(
		_config.SqliteDB,
		[]any{},
	)
	_server := server.Must(
		creds,
		_store,
	)
	lis, err := net.Listen("tcp", _config.GrpcConnect)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	logfile, err := os.OpenFile(_config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to set log file: %v", err)
	}
	defer logfile.Close()
	var writer = io.MultiWriter(os.Stdout, logfile)
	grpclog.SetLoggerV2(
		grpclog.NewLoggerV2WithVerbosity(
			writer, writer, writer,
			2,
		),
	)

	if err := _server.Serve(lis); err != nil {
		log.Fatalf("fatal server error: %v", err)
	}
	os.Exit(0)
}

func loadTLSCreds() (credentials.TransportCredentials, error) {
	cert, err := tls.X509KeyPair(TLSServerCert, TLSServerKey)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(TLSClientCert) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}
	if !certPool.AppendCertsFromPEM(TLSAdminCert) {
		return nil, fmt.Errorf("failed to add admin CA's certificate")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	return credentials.NewTLS(config), nil
}
