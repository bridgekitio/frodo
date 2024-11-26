package dismissal

import (
	"context"

	"github.com/bridgekit-io/frodo/fail"
)

type DismissServiceHandler struct{}

func (DismissServiceHandler) Dismiss(_ context.Context, req *DismissRequest) (*DismissResponse, error) {
	if req.Name == "" {
		return nil, fail.BadRequest("name is required")
	}
	return &DismissResponse{Value: "Goodbye, " + req.Name}, nil
}
