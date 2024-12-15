package api

import (
	"fmt"
	"io"
	"jamel/gen/go/jamel"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (a *Api) GetReport(id string) (*jamel.TaskResponse, error) {
	return a.client.GetReport(a.ctx, &jamel.ReportRequest{
		Id: id,
	})
}

func (a *Api) GetFile(id string, reportype jamel.ReportType) (string, error) {
	var (
		temp     = filepath.Join(uuid.NewString())
		filename string
		receive  int
		_p       int
	)

	stream, err := a.client.GetFile(a.ctx,
		&jamel.ReportRequest{
			Id:         id,
			ReportType: reportype,
		})
	if err != nil {
		return "", fmt.Errorf("error to get file stream: %w", err)
	}
	file, err := os.OpenFile(temp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer file.Close()
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			defer os.Remove(temp)
			defer file.Close()
			return "", err
		}
		filename = resp.Filename
		if _, err := file.Write(resp.Chunk); err != nil {
			return "", err
		}
		receive += len(resp.Chunk)
		percent := int(float64(receive) / float64(resp.Size) * 100)
		if (_p != percent) && (resp.Size > bufSize) && (percent <= 100) {
			fmt.Printf("\r⬅️ %s %d%%\r", filename, percent)
		}
		_p = percent
	}
	if err := os.Rename(temp, filepath.Join(filename)); err != nil {
		return "", fmt.Errorf("error rename: %w", err)
	}
	return filename, nil
}
