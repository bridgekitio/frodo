package dismissal

import "context"

type DismissService interface {
	Dismiss(ctx context.Context, req *DismissRequest) (*DismissResponse, error)
}

type DismissRequest struct {
	Name string
}

type DismissResponse struct {
	Value string
}
