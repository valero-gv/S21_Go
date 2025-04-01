package warehouse

import "errors"

var (
	ErrConfMismatch = errors.New("node configurations didn't match")
)
