package handler

import (
	"context"
)

const StaticDir = "static"

type Handler struct {
	ctx *context.Context
}

func New(
	_ctx *context.Context,
) *Handler {
	h := &Handler{
		ctx: _ctx,
	}
	return h
}
