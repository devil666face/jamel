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
		resp     = &jamel.TaskResponse{}
		temp     = filepath.Join(StaticDir, uuid.NewString())
	)
	file, err := os.OpenFile(temp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	defer os.Remove(temp)
	for {
		task, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("failed to recieve from stream: %w", err)
		}
		if _, err := file.Write(task.Chunk); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
		received += int64(len(task.Chunk))
		if received == task.Size {
			resp.TaskType = task.TaskType
			resp.Filename = task.Filename
			break
		}
	}

	resp.TaskId, err = h.s3.Upload(temp)
	if err != nil {
		return fmt.Errorf("failed to upload on s3: %w", err)
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		return fmt.Errorf("falied to serialize in json: %w", err)
	}
	if err := h.rmq.Publish(rmq.TaskQueue, data); err != nil {
		return fmt.Errorf("failed to set task in queue: %w", err)
	}

	resp, err = h.results.WaitResp(resp.TaskId, 120)
	if err != nil {
		return fmt.Errorf("failed to get resp from result queue: %w", err)
	}

	if err := h.manager.Task.Create(
		h.manager.Task.NewTask(resp),
	); err != nil {
		return fmt.Errorf("failed to write resp in database: %w", err)
	}

	return stream.SendAndClose(
		resp,
	)
}
