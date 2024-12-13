package api

import "jamel/gen/go/jamel"

func (a *Api) GetReport(id string) (*jamel.TaskResponse, error) {
	return a.client.GetReport(a.ctx, &jamel.ReportRequest{
		Id: id,
	})
}
