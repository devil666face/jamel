package server

import (
	"context"
	"encoding/json"
	"fmt"
	"jamel/gen/go/jamel"
	"jamel/internal/server/grpc/handler"
	"jamel/internal/server/service/models"
	"jamel/pkg/queue"
	"jamel/pkg/rmq"
	"log"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gorm.io/gorm"
)

type DB interface {
	DB() *gorm.DB
}

type Server struct {
	manager *models.Manager
	s3      handler.S3
	rmq     handler.Rmq
	resulst handler.Queue

	GrpcServer *grpc.Server

	jamel.UnimplementedJamelServiceServer
}

func Must(
	creds credentials.TransportCredentials,
	_db DB,
	_s3 handler.S3,
	_rmq handler.Rmq,
) *Server {
	_server := &Server{
		manager: models.New(_db.DB()),
		s3:      _s3,
		rmq:     _rmq,
		resulst: queue.New(),
	}
	_grpc := grpc.NewServer(
		grpc.Creds(creds),
	)
	jamel.RegisterJamelServiceServer(
		_grpc,
		_server,
	)
	_server.GrpcServer = _grpc
	return _server
}

func (s *Server) wrap(ctx *context.Context) *handler.Handler {
	return handler.New(
		ctx,
		s.manager,
		s.s3,
		s.rmq,
		s.resulst,
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

func (s *Server) NewTaskFromImage(ctx context.Context, request *jamel.TaskRequest) (*jamel.TaskResponse, error) {
	return s.
		wrap(&ctx).
		NewTaskFromImage(request)
}

func (s *Server) ResponseQueueHandler() error {
	var (
		respch      = make(chan amqp.Delivery)
		errch       = make(chan error)
		ctx, cancel = context.WithCancel(context.Background())
	)
	defer close(respch)
	defer close(errch)
	defer cancel()

	if err := s.rmq.Consume(ctx, rmq.ResultQueue, respch); err != nil {
		return fmt.Errorf("failed to consume queue: %w", err)
	}

	go func() {
		for data := range respch {
			var resp = jamel.TaskResponse{}
			log.Printf("recieved resp: %v", string(data.Body))
			if err := json.Unmarshal(data.Body, &resp); err != nil {
				errch <- fmt.Errorf("unmarshal resp from queue error: %w", err)
				continue
			}
			s.resulst.Set(&resp)

			switch resp.TaskType {
			case jamel.TaskType_DOCKER:
				continue
			default:
				if err := s.s3.Delete(resp.TaskId); err != nil {
					errch <- fmt.Errorf("failed to delete obj from s3: %w", err)
					continue
				}
			}
		}
	}()

	for err := range errch {
		if err != nil {
			// cancel()
			return fmt.Errorf("resp queue error: %w", err)
		}
	}
	return nil
}

func (s *Server) Reconnect() error {
	return s.rmq.Connect()
}
