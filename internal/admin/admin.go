package admin

import (
	"jamel/internal/admin/api"

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
