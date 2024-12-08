package handler

import (
	"context"
)

const StaticDir = "static"

type S3 interface {
	Upload(string) (string, error)
	Download(string) (string, error)
}

type Handler struct {
	ctx *context.Context
	s3  S3
}

func New(
	_ctx *context.Context,
	_s3 S3,

) *Handler {
	h := &Handler{
		ctx: _ctx,
		s3:  _s3,
	}
	return h
}
