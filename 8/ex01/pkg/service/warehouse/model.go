package warehouse

import (
	"time"

	"github.com/google/uuid"
)

type NodeInfo struct {
	Address string
	UUID    uuid.UUID
}

type FullNodeInfo struct {
	NodeInfo
	PulseInterval time.Duration
}
