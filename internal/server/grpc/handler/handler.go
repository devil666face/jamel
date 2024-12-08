package handler

import (
	"context"

	"github.com/streadway/amqp"
)

const StaticDir = "static"

type S3 interface {
	Upload(string) (string, error)
	Download(string) (string, error)
}

type Rmq interface {
	Publish(string, []byte) error
	Consume(string, chan<- amqp.Delivery) error
}

type Handler struct {
	ctx *context.Context
	s3  S3
	rmq Rmq
}

func New(
	_ctx *context.Context,
	_s3 S3,
	_rmq Rmq,
) *Handler {
	h := &Handler{
		ctx: _ctx,
		s3:  _s3,
		rmq: _rmq,
	}
	return h
}
