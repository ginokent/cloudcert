package controller

import (
	"context"

	"github.com/newtstat/cloudacme/proto/generated/go/v1/cloudacme"
	"github.com/newtstat/cloudacme/trace"
)

type TestAPIController struct {
	cloudacme.UnimplementedTestAPIServer
}

func (*TestAPIController) Echo(ctx context.Context, request *cloudacme.TestAPIEchoRequest) (response *cloudacme.TestAPIEchoResponse, err error) {
	_ = trace.StartFunc(ctx, "Echo")(func(ctx context.Context) error {
		response = &cloudacme.TestAPIEchoResponse{
			Message: request.GetMessage(),
		}

		return nil
	})

	return response, nil
}
