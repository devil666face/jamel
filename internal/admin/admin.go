package admin

import (
	"errors"
	"fmt"
	"jamel/gen/go/jamel"
	"jamel/internal/admin/api"
	"strings"

	"google.golang.org/grpc"
)

type Admin struct {
	Client *api.Api
}

func Must(
	username, passowrd string,
	conn *grpc.ClientConn,
) *Admin {
	return &Admin{
		Client: api.New(
			username, passowrd,
			conn,
		),
	}
}
func (a *Admin) NewTaskFromImage(image string) (string, error) {
	return a.newTask(func() (*jamel.TaskResponse, error) {
		return a.Client.NewTaskFromImage(image)
	})
}

func (a *Admin) NewTaskFromFile(filename string, tasktype jamel.TaskType) (string, error) {
	return a.newTask(func() (*jamel.TaskResponse, error) {
		return a.Client.NewTaskFromFile(filename, tasktype)
	})
}

func (a *Admin) newTask(taskFunc func() (*jamel.TaskResponse, error)) (string, error) {
	resp, err := taskFunc()
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return formatTaskResponse(resp), nil
}

func formatTaskResponse(resp *jamel.TaskResponse) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "\r%s\n%s\n%s\n", resp.TaskId, resp.Name, resp.Report)
	return sb.String()
}
