package admin

import (
	"jamel/internal/admin/api"

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

func (a *Admin) NewTaskFromFile(filename string) error {
	return a.client.NewTaskFromFile(filename)
}
