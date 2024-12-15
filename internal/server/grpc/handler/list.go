package handler

import (
	"fmt"
	"jamel/gen/go/jamel"
	"jamel/internal/server/service/models"
)

func (h *Handler) TaskList(request *jamel.Request) (*jamel.TaskListResponse, error) {
	var (
		resps = []*jamel.TaskResponse{}
	)
	tasks, err := h.manager.Task.All()
	if err != nil {
		return nil, fmt.Errorf("task list error: %w", err)
	}
	for _, task := range tasks {
		resps = append(resps,
			task.TaskToResp(func(t *models.Task) {
				t.Report = ""
				t.Json = ""
				t.Sbom = ""
			}),
		)
	}
	return &jamel.TaskListResponse{Tasks: resps}, nil
}
