package server

import (
	"context"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func Test_recoveryHandler(t *testing.T) {
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)
	ctx.SetUserContext(context.Background())
	type args struct {
		c *fiber.Ctx
		e interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "string",
			args: struct {
				c *fiber.Ctx
				e interface{}
			}{c: ctx, e: "error"},
		},
		{
			name: "error",
			args: struct {
				c *fiber.Ctx
				e interface{}
			}{c: ctx, e: errors.New("test")},
		},
		{
			name: "int",
			args: struct {
				c *fiber.Ctx
				e interface{}
			}{c: ctx, e: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recoveryHandler(tt.args.c, tt.args.e)
		})
	}
}
