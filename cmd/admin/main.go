package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"

	_ "embed"

	"jamel/internal/admin"
	"jamel/internal/admin/config"
	"jamel/internal/admin/view"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	force = flag.String("force", "", "use force for check image from docker hub without ui")
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
		log.Fatalf("failed to load server cert: %v\n", err)
	}
	_config := config.Must()
	conn, err := grpc.NewClient(
		_config.Server,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024),
			grpc.MaxCallSendMsgSize(10*1024*1024),
		),
	)
	if err != nil {
		log.Fatalf("error connect to server: %v\n", err)
	}
	defer conn.Close()
	_admin := admin.Must(
		_config.Username,
		_config.Password,
		conn,
	)
	_view := view.New(
		_admin,
	)

	flag.Parse()
	if *force != "" {
		out, err := _admin.Client.TaskFromImage(*force)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(out)
		os.Exit(0)
	}
	_view.Run()
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
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	}
	return credentials.NewTLS(config), nil
}
