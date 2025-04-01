package warehouse

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"slices"
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
	leader    uuid.UUID

	storage       Storage
	knownNodes    map[uuid.UUID]NodeConnInfo
	knownNodesMU  sync.RWMutex
	pulseInterval time.Duration
	hasher        *Hasher

	pulseStopCh chan struct{}
}

type Storage interface {
	Load(key uuid.UUID) ([]byte, bool)
	Store(key uuid.UUID, value []byte)
	Delete(key uuid.UUID) bool
	GetSize() int
}

// NodeConn describes an outgoing node connection.
type NodeConn interface {
	GetUUID() uuid.UUID
	GetAddress() string
	SendPulse(ctx context.Context, from NodeInfo) (FullNodeInfo, error)
	GetKnownNodes(ctx context.Context) ([]NodeConn, error)
	FixMesh(ctx context.Context, replicaFactor int)
	StorageDelete(ctx context.Context, key uuid.UUID) (bool, error)
	StorageLoad(ctx context.Context, key uuid.UUID) ([]byte, bool, error)
	StorageStore(ctx context.Context, key uuid.UUID, value []byte) error
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
	newID := uuid.New()

	result := &Node{
		myAddress:     myAddress,
		myID:          newID,
		leader:        newID,
		storage:       storage,
		knownNodes:    make(map[uuid.UUID]NodeConnInfo),
		knownNodesMU:  sync.RWMutex{},
		pulseInterval: pulseInterval,
		hasher:        nil,
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
	log.Println("Outgoing pulse for", len(me.knownNodes), "------ leader", me.leader)
	log.Println("Storage size", me.storage.GetSize())

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
		} else if connInfo.Conn.GetUUID() == me.leader {
			me.ElectNewLeader()
		}
	}

	known, _ := me.GetKnownNodes()
	if me.hasher != nil {
		err := me.hasher.UpdateOnlines(append(known, NodeInfo{Address: me.myAddress, UUID: me.myID}))
		if err != nil {
			log.Println(err)
		}
	}
}

func (me *Node) ElectNewLeader() {
	aliveConns, _ := me.GetKnownNodes()
	if len(aliveConns) < 1 {
		me.leader = me.myID

		return
	}

	aliveConns = append(aliveConns, NodeInfo{
		Address: me.myAddress,
		UUID:    me.myID,
	})

	slices.SortFunc(aliveConns, func(a, b NodeInfo) int {
		return bytes.Compare(a.UUID[:], b.UUID[:])
	})

	me.leader = aliveConns[0].UUID
}

// GetKnownNodes returns a list of the current valid connections.
func (me *Node) GetKnownNodes() ([]NodeInfo, NodeInfo) {
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

	return result, me.GetLeader()
}

func (me *Node) GetLeader() NodeInfo {
	me.knownNodesMU.RLock()
	defer me.knownNodesMU.RUnlock()

	if me.leader == me.myID {
		return NodeInfo{
			Address: me.myAddress,
			UUID:    me.myID,
		}
	}

	leaderConn := me.knownNodes[me.leader].Conn

	return NodeInfo{
		Address: leaderConn.GetAddress(),
		UUID:    leaderConn.GetUUID(),
	}
}

// IncomingPulse updates the LastSeen time of the connection.
func (me *Node) IncomingPulse(conn NodeConn) {
	log.Println("Incoming pulse from:", conn.GetUUID())

	pulseTime := time.Now()

	me.knownNodesMU.Lock()

	connUUID := conn.GetUUID()
	if _, ok := me.knownNodes[connUUID]; !ok && me.hasher != nil {
		log.Println("Didn't include a new node after we're fixed")

		me.knownNodesMU.Unlock()

		return
	}

	me.knownNodes[connUUID] = NodeConnInfo{
		Conn:      conn,
		LastPulse: pulseTime,
	}

	me.knownNodesMU.Unlock()

	// new promoted leader
	if bytes.Compare(connUUID[:], me.leader[:]) < 0 {
		me.leader = connUUID
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

func (me *Node) FixMesh(ctx context.Context, replicaFactor int) {
	if me.hasher != nil {
		return // already fixed
	}

	nodes := make([]NodeInfo, 0, len(me.knownNodes)+1)

	me.knownNodesMU.RLock()
	defer me.knownNodesMU.RUnlock()

	for _, connInfo := range me.knownNodes {
		if !connInfo.ThinkAlive(me.pulseInterval) {
			continue
		}

		nodes = append(nodes, NodeInfo{
			Address: connInfo.Conn.GetAddress(),
			UUID:    connInfo.Conn.GetUUID(),
		})
	}

	nodes = append(nodes, NodeInfo{
		Address: me.myAddress,
		UUID:    me.myID,
	})

	me.hasher = NewHasher(nodes, replicaFactor)

	for _, connInfo := range me.knownNodes {
		if !connInfo.ThinkAlive(me.pulseInterval) {
			continue
		}

		connInfo.Conn.FixMesh(ctx, replicaFactor)
	}
}

func (me *Node) ClientStorageLoad(ctx context.Context, key uuid.UUID) ([]byte, bool, error) {
	if me.hasher == nil {
		return nil, false, ErrNotFixed
	}

	if me.leader != me.myID {
		return nil, false, ErrNotALeader
	}

	targetNodes := me.hasher.KeyToNodes(key)
	if len(targetNodes) < 1 {
		return nil, false, ErrTargetNodesOffline
	}

	for _, node := range targetNodes {
		if node.UUID == me.myID {
			val, found := me.StorageLoad(key)
			if !found {
				continue
			}

			return val, found, nil
		}

		val, found, err := me.knownNodes[node.UUID].Conn.StorageLoad(ctx, key)
		if err != nil || !found {
			continue
		}

		return val, found, nil
	}

	return nil, false, ErrNotFound
}

func (me *Node) ClientStorageStore(ctx context.Context, key uuid.UUID, value []byte) (int, error) {
	if me.hasher == nil {
		return 0, ErrNotFixed
	}

	if me.leader != me.myID {
		return 0, ErrNotALeader
	}

	stored := 0

	targetNodes := me.hasher.KeyToNodes(key)
	if len(targetNodes) < 1 {
		return 0, ErrTargetNodesOffline
	}

	for _, node := range targetNodes {
		if node.UUID == me.myID {
			me.StorageStore(key, value)
			stored++

			continue
		}

		err := me.knownNodes[node.UUID].Conn.StorageStore(ctx, key, value)
		if err != nil {
			continue
		}

		stored++
	}

	return stored, nil
}

func (me *Node) ClientStorageDelete(ctx context.Context, key uuid.UUID) (int, error) {
	if me.hasher == nil {
		return 0, ErrNotFixed
	}

	if me.leader != me.myID {
		return 0, ErrNotALeader
	}

	deleted := 0

	targetNodes := me.hasher.KeyToNodes(key)
	if len(targetNodes) < 1 {
		return 0, ErrTargetNodesOffline
	}

	for _, node := range targetNodes {
		if node.UUID == me.myID {
			found := me.StorageDelete(key)
			if found {
				deleted++
			}

			continue
		}

		haveDeleted, err := me.knownNodes[node.UUID].Conn.StorageDelete(ctx, key)
		if err != nil || !haveDeleted {
			continue
		}

		deleted++
	}

	return deleted, nil
}
