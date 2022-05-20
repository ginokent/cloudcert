package controller_test

import (
	"context"
	"testing"

	"github.com/ginokent/cloudacme/controller"
	"github.com/ginokent/cloudacme/proto/generated/go/v1/cloudacme"
)

func TestTestAPIController_Echo(t *testing.T) {
	t.Parallel()
	t.Run("success()", func(t *testing.T) {
		t.Parallel()
		tr := &controller.TestAPIController{}
		const wantMessage = "test"
		gotResponse, err := tr.Echo(context.Background(), &cloudacme.TestAPIEchoRequest{Message: wantMessage})
		if err != nil {
			t.Errorf("TestAPIController.Echo() error = %v, wantErr %v", err, nil)
			return
		}
		if gotResponse.Message != wantMessage {
			t.Errorf("TestAPIController.Echo(): Message = %v, want %v", gotResponse.Message, wantMessage)
		}
	})
}
