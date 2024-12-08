package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"jamel/gen/go/jamel"
	"jamel/pkg/fs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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

func (a *Api) NewTaskFromFile(filename string) error {
	var (
		sent int
		_p   int
	)
	file, stat, err := fs.OpenFile(filename)
	if err != nil {
		return fmt.Errorf("error to send grpc upload file: %w", err)
	}
	defer file.Close()
	stream, err := a.client.NewTaskFromFile(a.ctx)
	if err != nil {
		return fmt.Errorf("error to start upload stream: %w", err)
	}
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			defer stream.CloseSend()
			return fmt.Errorf("error to read file chunk: %w", err)
		}
		if err := stream.Send(&jamel.TaskRequest{
			Filename: file.Name(),
			Size:     stat.Size(),
			Chunk:    buf[:n],
		}); err != nil {
			return fmt.Errorf("error to send file chunk via grpc: %w", err)
		}
		sent += len(buf)
		percent := int(float64(sent) / float64(stat.Size()) * 100)
		if _p != percent {
			fmt.Printf("uploading %s, transferred %d%%\n", file.Name(), percent)
		}
		_p = percent
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("error to get success request about uploading: %w", err)
	}
	fmt.Println(resp)
	return nil
}
