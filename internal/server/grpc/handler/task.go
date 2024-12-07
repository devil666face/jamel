package handler

import (
	"fmt"
	"os"
	"path/filepath"

	"jamel/gen/go/jamel"

	"github.com/google/uuid"
)

func (h *Handler) NewTaskFromFile(stream jamel.JamelService_NewTaskFromFileServer) error {
	var (
		received int64
		temp     = filepath.Join(StaticDir, uuid.NewString())
	)
	file, err := os.OpenFile(temp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	for {
		resp, err := stream.Recv()
		if err != nil {
			defer os.Remove(temp)
			defer file.Close()
			return fmt.Errorf("failed to recieve from stream: %w", err)
		}
		if _, err := file.Write(resp.Chunk); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
		received += int64(len(resp.Chunk))
		if received == resp.Size {
			if err := os.Rename(temp, filepath.Join(StaticDir, resp.Filename)); err != nil {
				return fmt.Errorf("failed rename %s to %s", temp, resp.Filename)
			}
			return stream.SendAndClose(
				&jamel.TaskResponse{
					TaskId: uuid.NewString(),
				},
			)
		}
	}
}
