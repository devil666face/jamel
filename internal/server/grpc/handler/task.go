package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"jamel/gen/go/jamel"
	"jamel/pkg/rmq"

	"github.com/google/uuid"
)

func (h *Handler) NewTaskFromFile(stream jamel.JamelService_NewTaskFromFileServer) error {
	var (
		received int64
		filename string
		temp     = filepath.Join(StaticDir, uuid.NewString())
	)
	file, err := os.OpenFile(temp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	defer os.Remove(temp)
	for {
		resp, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("failed to recieve from stream: %w", err)
		}
		if _, err := file.Write(resp.Chunk); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
		received += int64(len(resp.Chunk))
		if received == resp.Size {
			filename = resp.Filename
			break
		}
	}

	objid, err := h.s3.Upload(temp)
	if err != nil {
		return fmt.Errorf("failed to upload on s3: %w", err)
	}

	var resp = &jamel.TaskResponse{
		TaskId:   objid,
		Filename: filename,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("falied to serialize in json: %w", err)
	}
	if err := h.rmq.Publish(rmq.TaskQueue, data); err != nil {
		return fmt.Errorf("failed to set task in queue: %w", err)
	}
	return stream.SendAndClose(
		resp,
	)
}
