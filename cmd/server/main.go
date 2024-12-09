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
	"jamel/pkg/rmq"
	"jamel/pkg/s3"

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
	if err := os.MkdirAll("db", 0o755); err != nil {
		log.Fatalf("failed to create dir for db: %v", err)
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
	_s3, err := s3.New(
		_config.S3Conntect,
		_config.S3Username, _config.S3Password,
		_config.S3Bucket,
	)
	if err != nil {
		log.Fatalf("failed to create s3 obj: %v", err)
	}

	_rmq, err := rmq.New(
		_config.RMQConnect,
		_config.RMQUsername, _config.RMQPassword,
		rmq.TaskQueue, rmq.ResultQueue,
	)
	if err != nil {
		log.Fatalf("failed rmq connect: %v", err)
	}

	_server := server.Must(
		creds,
		_store,
		_s3,
		_rmq,
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

	go func() {
		for {
			if err := _server.ResponseQueueHandler(); err != nil {
				log.Printf("critical queue runtime error: %v", err)
			}
		}
	}()

	if err := _server.GrpcServer.Serve(lis); err != nil {
		log.Fatalf("fatal server error: %v", err)
	}
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
