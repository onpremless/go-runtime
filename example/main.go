package main

import (
	"context"
	"io"

	opl "github.com/onpremless/go-runtime/latest"
)

func main() {
	opl.Lambda(func(ctx context.Context, req *opl.Request[io.Reader]) (int, interface{}) {
		return 200, req.Payload
	})
}
