package warehouse

import "errors"

var (
	ErrNewNode            = errors.New("a new node was connected to the network after the initialization")
	ErrConfMismatch       = errors.New("node configurations didn't match")
	ErrNotALeader         = errors.New("address your request to the leader")
	ErrNotFixed           = errors.New("the mesh is not fixed yet")
	ErrNotFound           = errors.New("key not found in the nodes")
	ErrTargetNodesOffline = errors.New("all hashed node are offline")
)
