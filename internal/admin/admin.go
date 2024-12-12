package admin

import (
	"fmt"
	"jamel/gen/go/jamel"
	"jamel/internal/admin/api"
	"strings"

	"google.golang.org/grpc"
)

type Admin struct {
	client *api.Api
}

func Must(
	username, passowrd string,
	conn *grpc.ClientConn,
) *Admin {
	return &Admin{
		client: api.New(
			username, passowrd,
			conn,
		),
	}
}

func (a *Admin) NewTaskFromFile(filename string, tasktype jamel.TaskType) (string, error) {
	resp, err := a.client.NewTaskFromFile(filename, tasktype)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintln(resp.TaskId))
	sb.WriteString(fmt.Sprintln(resp.Filename))
	sb.WriteString(fmt.Sprintln(resp.Report))
	return sb.String(), nil
}
