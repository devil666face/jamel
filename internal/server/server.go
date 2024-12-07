package server

import (
	"context"
	"jamel/gen/go/jamel"
	"jamel/internal/server/grpc/handler"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gorm.io/gorm"
)

type DB interface {
	DB() *gorm.DB
}

type Server struct {
	db DB
	jamel.UnimplementedJamelServiceServer
}

func Must(
	creds credentials.TransportCredentials,
	_db DB,
) *grpc.Server {
	_server := &Server{
		db: _db,
	}
	_grpc := grpc.NewServer(
		grpc.Creds(creds),
	)
	jamel.RegisterJamelServiceServer(
		_grpc,
		_server,
	)
	return _grpc
}

func (s *Server) wrap(ctx *context.Context) *handler.Handler {
	return handler.New(
		ctx,
	)
}

type Contextable interface {
	Context() context.Context
}

func (s *Server) streamwrap(stream Contextable) *handler.Handler {
	var ctx = stream.Context()
	return s.wrap(&ctx)
}

func (s *Server) NewTaskFromFile(stream jamel.JamelService_NewTaskFromFileServer) error {
	return s.
		streamwrap(stream).
		NewTaskFromFile(stream)
}
