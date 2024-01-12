package sensitive

import (
	"context"
)

/*
 * Yes, the logic in these methods is garbage. The important stuff comes from the authorization/middleware that
 * determines if these methods even get called in the first place.
 */

type SecretServiceHandler struct{}

func (s *SecretServiceHandler) GetGroup(_ context.Context, req *GetGroupRequest) (*GroupResponse, error) {
	return &GroupResponse{ID: req.ID, Name: "Group " + req.ID}, nil
}

func (s *SecretServiceHandler) RenameGroup(_ context.Context, req *RenameGroupRequest) (*GroupResponse, error) {
	return &GroupResponse{ID: req.ID, Name: req.Name}, nil
}
