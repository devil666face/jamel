package api

import "jamel/gen/go/jamel"

func (a *Api) TaskList() (*jamel.TaskListResponse, error) {
	return a.client.TaskList(a.ctx, &jamel.Request{})
}
