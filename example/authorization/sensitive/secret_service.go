package sensitive

import "context"

// SecretService contains some contrived operations to show you how to lock down your endpoints using
// the Authorization metadata helpers and the ROLES doc option.
type SecretService interface {
	// GetGroup looks up a group given its unique id.
	//
	// GET    /group/{ID}
	// ROLES  admin.read, group.{ID}.read
	GetGroup(context.Context, *GetGroupRequest) (*GroupResponse, error)

	// RenameGroup changes an existing group's name.
	//
	// PUT    /group/{ID}/name
	// ROLES  admin.write, group.{ID}.write
	RenameGroup(context.Context, *RenameGroupRequest) (*GroupResponse, error)
}

type GetGroupRequest struct {
	ID string
}

type RenameGroupRequest struct {
	ID   string
	Name string
}

type GroupResponse struct {
	ID   string
	Name string
}
