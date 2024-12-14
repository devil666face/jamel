package api

import (
	"fmt"
	"io"
	"jamel/gen/go/jamel"
	"jamel/pkg/fs"
)

func (a *Api) TaskFromImage(image string) (*jamel.TaskResponse, error) {
	return a.client.TaskFromImage(a.ctx, &jamel.TaskRequest{
		Filename: image,
		TaskType: jamel.TaskType_DOCKER,
	})
}

func (a *Api) TaskFromFile(filename string, tasktype jamel.TaskType) (*jamel.TaskResponse, error) {
	var (
		sent int
		_p   int
	)
	file, stat, err := fs.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error to send grpc upload file: %w", err)
	}
	defer file.Close()
	stream, err := a.client.TaskFromFile(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("error to start upload stream: %w", err)
	}
	buf := make([]byte, bufSize)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			defer stream.CloseSend()
			return nil, fmt.Errorf("error to read file chunk: %w", err)
		}
		if err := stream.Send(&jamel.TaskRequest{
			Filename: file.Name(),
			Size:     stat.Size(),
			Chunk:    buf[:n],
			TaskType: tasktype,
		}); err != nil {
			return nil, fmt.Errorf("error to send file chunk via grpc: %w", err)
		}
		sent += len(buf)
		percent := int(float64(sent) / float64(stat.Size()) * 100)
		if (_p != percent) && (stat.Size() > bufSize) {
			fmt.Printf("\r➡️ %s %d%%\r", file.Name(), percent)
		}
		_p = percent
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("error to get success request about uploading: %w", err)
	}
	return resp, err
}
