package handler

import (
	"context"
	"jamel/gen/go/jamel"
	"jamel/internal/server/service/models"
	"time"

	"github.com/streadway/amqp"
)

const StaticDir = "static"

type S3 interface {
	Upload(string) (string, error)
	Download(string) (string, error)
	Delete(string) error
}

type Rmq interface {
	Publish(string, []byte) error
	Consume(context.Context, string, chan<- amqp.Delivery) error
	Connect() error
}

type Queue interface {
	Get(string) (*jamel.TaskResponse, error)
	Set(*jamel.TaskResponse)
	WaitResp(string, ...time.Duration) (*jamel.TaskResponse, error)
}

type Handler struct {
	ctx     *context.Context
	manager *models.Manager
	s3      S3
	rmq     Rmq
	results Queue
}

func New(
	_ctx *context.Context,
	_manager *models.Manager,
	_s3 S3,
	_rmq Rmq,
	_results Queue,
) *Handler {
	h := &Handler{
		ctx:     _ctx,
		manager: _manager,
		s3:      _s3,
		rmq:     _rmq,
		results: _results,
	}
	return h
}
