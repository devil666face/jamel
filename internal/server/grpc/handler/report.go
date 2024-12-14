package handler

import (
	"bytes"
	"fmt"
	"io"
	"jamel/gen/go/jamel"
	"jamel/internal/server/service/models"
)

func (h *Handler) GetReport(request *jamel.ReportRequest) (*jamel.TaskResponse, error) {
	task, err := h.manager.Task.Get(request.Id)
	if err != nil {
		return nil, fmt.Errorf("report error: %w", err)
	}
	return task.TaskToResp(
		func(t *models.Task) { t.Json = ""; t.Sbom = "" },
	), nil
}

func (h *Handler) GetFile(request *jamel.ReportRequest, stream jamel.JamelService_GetFileServer) error {
	task, err := h.manager.Task.Get(request.Id)
	if err != nil {
		return fmt.Errorf("report error: %w", err)
	}

	var (
		name   string
		reader *bytes.Reader
	)
	switch request.ReportType {
	case jamel.ReportType_JSON:
		name = fmt.Sprintf("%s.json", task.Name)
		reader = bytes.NewReader([]byte(task.Json))
	case jamel.ReportType_SBOM_R:
		name = fmt.Sprintf("%s.sbom.json", task.Name)
		reader = bytes.NewReader([]byte(task.Sbom))
	case jamel.ReportType_PDF:
		return fmt.Errorf("unimplement")
	}

	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error reading file chunk: %w", err)
		}

		if err := stream.Send(&jamel.FileResponse{
			Filename: name,
			Size:     int64(reader.Len()),
			Chunk:    buf[:n],
		}); err != nil {
			return fmt.Errorf("error sending file chunk via grpc: %w", err)
		}
	}
}
