package handler

import (
	"fmt"
	"jamel/gen/go/jamel"
	"jamel/internal/server/service/models"
)

func (h *Handler) GetReport(request *jamel.ReportRequest) (*jamel.TaskResponse, error) {
	fmt.Println(request.Id)
	task, err := h.manager.Task.Get(request.Id)
	if err != nil {
		return nil, fmt.Errorf("report error: %w", err)
	}
	return task.TaskToResp(
		func(t *models.Task) { t.Json = ""; t.Sbom = "" },
	), nil
}
