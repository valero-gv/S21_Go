package warehouse

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	knownNodesGroomerFactor = 15 // every X pulse intervals the known hosts are groomed down.
	thinkAliveFactor        = 3  // if last aknowledge was more than X pulse intervals ago, count it dead.
)

// Node is a node object.
type Node struct {
	myID      uuid.UUID
	myAddress string

	storage       Storage
	knownNodes    map[uuid.UUID]NodeConnInfo
	knownNodesMU  sync.RWMutex
	pulseInterval time.Duration

	pulseStopCh chan struct{}
}

type Storage interface {
	Load(key uuid.UUID) ([]byte, bool)
	Store(key uuid.UUID, value []byte)
	Delete(key uuid.UUID) bool
}

// NodeConn describes an outgoing node connection.
type NodeConn interface {
	GetUUID() uuid.UUID
	GetAddress() string
	SendPulse(ctx context.Context, from NodeInfo) (FullNodeInfo, error)
	GetKnownNodes(ctx context.Context) ([]NodeConn, error)
}

// NodeConnInfo describes an alive or dead outgoing connection.
type NodeConnInfo struct {
	Conn      NodeConn
	LastPulse time.Time
}

// ThinkAlive checks if a connection is still considered alive.
func (info NodeConnInfo) ThinkAlive(pulseInterval time.Duration) bool {
	return info.LastPulse.Add(thinkAliveFactor * pulseInterval).After(time.Now())
}

// NewNode creates a new node.
func NewNode(ctx context.Context, myAddress string, pulseInterval time.Duration, storage Storage) *Node {
	result := &Node{
		myAddress:     myAddress,
		myID:          uuid.New(),
		storage:       storage,
		knownNodes:    make(map[uuid.UUID]NodeConnInfo),
		knownNodesMU:  sync.RWMutex{},
		pulseInterval: pulseInterval,
		pulseStopCh:   nil, // nillness indicates that the pulsating is not still running.
	}

	result.startOutgoingPulse(ctx)

	return result
}

// GetPulseInterval returns the pulse interval of the node.
func (me *Node) GetPulseInterval() time.Duration {
	return me.pulseInterval
}

func (me *Node) GetAddress() string {
	return me.myAddress
}

// GetNodeUUID returns the uuid of the node.
func (me *Node) GetNodeUUID() uuid.UUID {
	return me.myID
}

func (me *Node) startOutgoingPulse(ctx context.Context) {
	if me.pulseStopCh != nil { // already pulsating
		return
	}

	me.pulseStopCh = make(chan struct{})

	go func() {
		pulseTicker := time.NewTicker(me.pulseInterval)
		groomerTicker := time.NewTicker(knownNodesGroomerFactor * me.pulseInterval)

		for {
			select {
			case <-groomerTicker.C:
				me.GroomKnownHosts()
			case <-pulseTicker.C:
				me.OutgoingPulse(ctx)
			case <-me.pulseStopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// OutgoingPulse forces the connections to send a pulse.
func (me *Node) OutgoingPulse(ctx context.Context) {
	log.Println("Outgoing pulse for", len(me.knownNodes))

	me.knownNodesMU.RLock()
	defer me.knownNodesMU.RUnlock()

	for _, connInfo := range me.knownNodes {
		if connInfo.ThinkAlive(me.pulseInterval) {
			// the uuids are immutable, so if the initial pulse
			// has resulted in no error, other errors don't really matter.
			go func(conn NodeConnInfo) {
				_, err := conn.Conn.SendPulse(
					ctx,
					NodeInfo{Address: me.myAddress, UUID: me.myID},
				)
				if err != nil {
					log.Println("Error Send pulse:", err.Error())
				}
			}(connInfo)
		}
	}
}

// GetKnownNodes returns a list of the current valid connections.
func (me *Node) GetKnownNodes() []NodeInfo {
	result := make([]NodeInfo, 0, len(me.knownNodes))

	me.knownNodesMU.RLock()
	defer me.knownNodesMU.RUnlock()

	for _, connInfo := range me.knownNodes {
		if !connInfo.ThinkAlive(me.pulseInterval) {
			continue
		}

		result = append(result, NodeInfo{
			Address: connInfo.Conn.GetAddress(),
			UUID:    connInfo.Conn.GetUUID(),
		})
	}

	return result
}

// IncomingPulse updates the LastSeen time of the connection.
func (me *Node) IncomingPulse(conn NodeConn) {
	log.Println("Incoming pulse from:", conn.GetUUID())

	pulseTime := time.Now()

	me.knownNodesMU.Lock()
	defer me.knownNodesMU.Unlock()

	me.knownNodes[conn.GetUUID()] = NodeConnInfo{
		Conn:      conn,
		LastPulse: pulseTime,
	}
}

// GroomKnownHosts frees memory occupied by the disconnected nodes.
func (me *Node) GroomKnownHosts() {
	me.knownNodesMU.Lock()
	defer me.knownNodesMU.Unlock()

	for _, connInfo := range me.knownNodes {
		if !connInfo.ThinkAlive(me.pulseInterval) {
			delete(me.knownNodes, connInfo.Conn.GetUUID())
		}
	}
}

// AddKnown is supposed to add a known node into the connections.
func (me *Node) AddKnown(ctx context.Context, conn NodeConn) error {
	resp, err := conn.SendPulse(ctx, NodeInfo{Address: me.myAddress, UUID: me.myID})
	if err != nil {
		return fmt.Errorf("initial pulse: %w", err)
	}

	if resp.PulseInterval != me.pulseInterval || resp.UUID != conn.GetUUID() {
		return ErrConfMismatch
	}

	conns, err := conn.GetKnownNodes(ctx)
	if err != nil {
		return fmt.Errorf("get known nodes: %w", err)
	}

	for _, newConn := range conns {
		if newConn.GetUUID() != me.GetNodeUUID() {
			me.IncomingPulse(newConn)
		}
	}

	return nil
}

func (me *Node) StorageLoad(key uuid.UUID) ([]byte, bool) {
	return me.storage.Load(key)
}

func (me *Node) StorageStore(key uuid.UUID, value []byte) {
	me.storage.Store(key, value)
}

func (me *Node) StorageDelete(key uuid.UUID) bool {
	return me.storage.Delete(key)
}
