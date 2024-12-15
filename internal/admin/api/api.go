package api

import (
	"context"
	"encoding/base64"
	"jamel/gen/go/jamel"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const bufSize = 1024

type Api struct {
	md     metadata.MD
	ctx    context.Context
	client jamel.JamelServiceClient
}

func New(
	username, password string,
	conn *grpc.ClientConn,
) *Api {
	var (
		creds = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		md    = metadata.Pairs("authorization", "Basic "+creds)
	)

	return &Api{
		md:     md,
		ctx:    metadata.NewOutgoingContext(context.Background(), md),
		client: jamel.NewJamelServiceClient(conn),
	}
}
