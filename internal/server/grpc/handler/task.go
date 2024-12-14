package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"jamel/gen/go/jamel"
	"jamel/internal/server/service/models"
	"jamel/pkg/rmq"

	"github.com/google/uuid"
)

const errorTimeout = 180

func (h *Handler) TaskFromImage(request *jamel.TaskRequest) (*jamel.TaskResponse, error) {
	var resp = &jamel.TaskResponse{
		TaskId:   uuid.NewString(),
		Name:     request.Filename,
		TaskType: jamel.TaskType_DOCKER,
	}
	data, err := json.Marshal(&resp)
	if err != nil {
		return nil, fmt.Errorf("falied to serialize in json: %w", err)
	}
	if err := h.rmq.Publish(rmq.TaskQueue, data); err != nil {
		return nil, fmt.Errorf("failed to set task in queue: %w", err)
	}

	resp, err = h.results.WaitResp(resp.TaskId, errorTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to get resp from result queue: %w", err)
	}

	if err := h.manager.Task.Create(
		h.manager.Task.NewTask(resp, func(t *models.Task) { resp.Json = ""; resp.Sbom = "" }),
	); err != nil {
		return nil, fmt.Errorf("failed to write resp in database: %w", err)
	}
	return resp, nil
}

func (h *Handler) TaskFromFile(stream jamel.JamelService_TaskFromFileServer) error {
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
			resp.Name = task.Filename
			break
		}
	}

	resp.TaskId, err = h.s3.Upload(temp)
	if err != nil {
		return fmt.Errorf("failed to upload on s3: %w", err)
	}
	go os.Remove(temp)

	data, err := json.Marshal(&resp)
	if err != nil {
		return fmt.Errorf("falied to serialize in json: %w", err)
	}
	if err := h.rmq.Publish(rmq.TaskQueue, data); err != nil {
		return fmt.Errorf("failed to set task in queue: %w", err)
	}

	resp, err = h.results.WaitResp(resp.TaskId, errorTimeout)
	if err != nil {
		return fmt.Errorf("failed to get resp from result queue: %w", err)
	}

	if err := h.manager.Task.Create(
		h.manager.Task.NewTask(resp, func(t *models.Task) { resp.Json = ""; resp.Sbom = "" }),
	); err != nil {
		return fmt.Errorf("failed to write resp in database: %w", err)
	}
	return stream.SendAndClose(
		resp,
	)
}
