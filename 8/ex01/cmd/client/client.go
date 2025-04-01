package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex01/pkg/service/api/warehousepb"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex01/pkg/service/warehouse"
)

type appConf struct {
	initialNodeAddress string
	initialNodeUUID    uuid.UUID
	replicaFactor      int
	pulseInterval      time.Duration
}

type HasherInterface interface {
	KeyToNodes(key uuid.UUID) []warehouse.NodeInfo
	UpdateOnlines(nodes []warehouse.NodeInfo) error
}

type OnlineNode struct {
	warehouse.NodeInfo
	Online bool
}

type Hasher struct {
	knownNodes        map[uuid.UUID]int
	knownNodesAddress []OnlineNode
	replicaFactor     int
}

type CommandType string

const (
	GetCommand    CommandType = "GET"
	SetCommand    CommandType = "SET"
	DeleteCommand CommandType = "DELETE"
)

var (
	ErrNewNode            = errors.New("a new node was connected to the network after the initialization")
	ErrKnownNodeFormat    = errors.New("known node wrong format")
	ErrReplicaFactor      = errors.New("this replica factor is not allowed")
	ErrNoMoreNodes        = errors.New("mesh ran out of nodes")
	ErrTargetNodesOffline = errors.New("all hashed node are offline")
	ErrValueNotFound      = errors.New("value not found")
)

func main() {
	appCtx, appCtxCancel := context.WithCancel(context.Background())
	defer appCtxCancel()

	conf, err := readConf()
	if err != nil {
		log.Fatal("Err Read conf: ", err.Error())
	}

	err = startREPL(appCtx, conf)
	if err != nil {
		log.Printf("Err Run repl: %s\n", err)
	}
}

func readConf() (appConf, error) {
	flagReplicaFactor := flag.Int("r", 0, "replica factor")
	flagPulseInterval := flag.String("p", "", "pulse period")
	flagKnownInfo := flag.String("kn", "", "known host info")

	flag.Parse()

	if *flagReplicaFactor < 1 {
		return appConf{}, ErrReplicaFactor
	}

	pulseInterval, err := time.ParseDuration(*flagPulseInterval)
	if err != nil {
		return appConf{}, fmt.Errorf("parse pulse interval: %w", err)
	}

	known := strings.TrimSpace(*flagKnownInfo)
	knownInfo := strings.Split(known, " ")

	if known == "" || len(knownInfo) != 2 {
		return appConf{}, ErrKnownNodeFormat
	}

	knownNodeID, err := uuid.Parse(knownInfo[1])
	if err != nil {
		return appConf{}, fmt.Errorf("known parse node uuid: %w", err)
	}

	return appConf{
		pulseInterval:      pulseInterval,
		initialNodeAddress: knownInfo[0],
		initialNodeUUID:    knownNodeID,
		replicaFactor:      *flagReplicaFactor,
	}, nil
}

func startREPL(ctx context.Context, conf appConf) error {
	conn, err := grpc.NewClient(conf.initialNodeAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("initial new client conn: %w", err)
	}

	nodes, err := getNodes(ctx, conn)
	if err != nil {
		return fmt.Errorf("get nodes: %w", err)
	}

	hasher := NewHasher(nodes, conf.replicaFactor)

	err = repl(ctx, conf.replicaFactor, conf.pulseInterval, conn, nodes, hasher)
	if err != nil {
		return fmt.Errorf("do repl: %w", err)
	}

	return nil
}

func getNodes(ctx context.Context, conn *grpc.ClientConn) ([]warehouse.NodeInfo, error) {
	mesh, err := warehousepb.NewWarehouseClient(conn).GetKnownNodes(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("make request get nodes: %w", err)
	}

	nodes, err := convertMesh(mesh.GetNodes())
	if err != nil {
		return nil, fmt.Errorf("convert mesh nodes: %w", err)
	}

	return nodes, nil
}

func convertMesh(nodes []*warehousepb.Node) ([]warehouse.NodeInfo, error) {
	myNodes := make([]warehouse.NodeInfo, 0, len(nodes))

	for _, node := range nodes {
		nodeID, err := uuid.Parse(node.GetNodeUuid())
		if err != nil {
			return nil, fmt.Errorf("parse node id: %w", err)
		}

		myNodes = append(myNodes, warehouse.NodeInfo{
			Address: node.GetAddress(),
			UUID:    nodeID,
		})
	}

	return myNodes, nil
}

func repl(ctx context.Context, replicaFactor int, pulseInterval time.Duration, conn *grpc.ClientConn,
	nodes []warehouse.NodeInfo, hasher HasherInterface,
) error {
	inpChan := make(chan string)

	go func(ch chan string) {
		reader := bufio.NewReader(os.Stdin)

		for {
			log.Println(">")

			s, err := reader.ReadString('\n')
			if err != nil {
				close(ch)

				return
			}
			ch <- s
		}
	}(inpChan)

stdinloop:
	for {
		select {
		case <-ctx.Done():
			return nil
		case stdin, ok := <-inpChan:
			if !ok {
				break stdinloop
			}
			err := handleCommand(ctx, stdin, hasher)
			if err != nil {
				log.Println("Err Handle command:", err)
			}
		case <-time.After(pulseInterval):
			newNodes, err := getNodes(ctx, conn)
			if err != nil {
				conn, newNodes = getNewConn(ctx, nodes)
			}
			if conn == nil {
				return ErrNoMoreNodes
			}
			nodes = newNodes

			err = hasher.UpdateOnlines(nodes)
			if err != nil {
				return fmt.Errorf("update nodes statuses: %w", err)
			}

			if len(nodes) < replicaFactor {
				log.Println("mesh size is less than the set replica factor")
			}
		}
	}

	return nil
}

func handleCommand(ctx context.Context, command string, hasher HasherInterface) error {
	commandType, key, value, err := parseCommand(command)
	if err != nil {
		return fmt.Errorf("parse command: %w", err)
	}

	targetNodes := hasher.KeyToNodes(key)
	conn := (*grpc.ClientConn)(nil)

	if len(targetNodes) < 1 {
		return ErrTargetNodesOffline
	}

	switch commandType {
	case SetCommand:
		saved := 0

		for _, currNode := range targetNodes {
			conn, err = grpc.NewClient(currNode.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("new conn to %s: %s\n", currNode.UUID.String(), err)

				continue
			}

			_, err = warehousepb.NewWarehouseClient(conn).StorageStore(ctx, &warehousepb.StorageStoreRequest{
				Key:   key.String(),
				Value: *value,
			})
			if err != nil {
				log.Printf("set value at %s: %s\n", currNode.UUID.String(), err)

				continue
			}

			saved++
		}

		log.Printf("Saved %d replicas\n", saved)
	case GetCommand:
		for _, currNode := range targetNodes {
			conn, err = grpc.NewClient(currNode.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("new conn to %s: %s\n", currNode.UUID.String(), err)

				continue
			}

			var resp *warehousepb.StorageLoadResponse

			resp, err = warehousepb.NewWarehouseClient(conn).StorageLoad(ctx, &warehousepb.StorageLoadRequest{
				Key: key.String(),
			})
			if err != nil {
				log.Printf("get value at %s: %s\n", currNode.UUID.String(), err)

				continue
			}

			if !resp.GetFound() {
				continue
			}

			log.Println(resp.GetValue())

			return nil
		}

		return ErrValueNotFound
	case DeleteCommand:
		deleted := 0

		for _, currNode := range targetNodes {
			conn, err = grpc.NewClient(currNode.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("new conn to %s: %s\n", currNode.UUID.String(), err)

				continue
			}

			var resp *warehousepb.StorageDeleteResponse

			resp, err = warehousepb.NewWarehouseClient(conn).StorageDelete(ctx, &warehousepb.StorageDeleteRequest{
				Key: key.String(),
			})
			if err != nil {
				log.Printf("delete value at %s: %s\n", currNode.UUID.String(), err)

				continue
			}

			if !resp.GetFound() {
				continue
			}

			deleted++
		}

		log.Printf("Delted %d replicas\n", deleted)
	}

	return nil
}

func NewHasher(initNodes []warehouse.NodeInfo, replicaFactor int) *Hasher {
	result := &Hasher{
		knownNodes:        make(map[uuid.UUID]int, len(initNodes)),
		knownNodesAddress: make([]OnlineNode, 0, len(initNodes)),
		replicaFactor:     replicaFactor,
	}

	if result.replicaFactor > len(initNodes) {
		result.replicaFactor = len(initNodes)
	}

	slices.SortFunc(initNodes, func(a, b warehouse.NodeInfo) int {
		return bytes.Compare(a.UUID[:], b.UUID[:])
	})

	for i, node := range initNodes {
		result.knownNodesAddress = append(result.knownNodesAddress, OnlineNode{
			Online:   true,
			NodeInfo: node,
		})

		result.knownNodes[node.UUID] = i
	}

	return result
}

func (h Hasher) UpdateOnlines(nodes []warehouse.NodeInfo) error {
	for i := range h.knownNodesAddress {
		h.knownNodesAddress[i].Online = false
	}

	for _, node := range nodes {
		idx, ok := h.knownNodes[node.UUID]
		if !ok {
			return ErrNewNode
		}

		h.knownNodesAddress[idx].Online = true
	}

	return nil
}

func (h Hasher) KeyToNodes(key uuid.UUID) []warehouse.NodeInfo {
	nodeIndices := make([]warehouse.NodeInfo, 0, h.replicaFactor)
	hash := fnv.New64a()
	keyBytes := []byte(key.String())

	for replicaIndex := range h.replicaFactor {
		hash.Reset()

		_, err := hash.Write(keyBytes)
		if err != nil {
			continue
		}

		baseHash := hash.Sum64()
		replicaHash := baseHash + uint64(replicaIndex)
		nodeIndex := int(replicaHash % uint64(len(h.knownNodesAddress)))

		if h.knownNodesAddress[nodeIndex].Online {
			nodeIndices = append(nodeIndices, h.knownNodesAddress[nodeIndex].NodeInfo)
		}
	}

	return nodeIndices
}

func parseCommand(line string) (CommandType, uuid.UUID, *string, error) {
	line = strings.TrimSpace(line)

	if line == "" {
		return "", uuid.Nil, nil, errors.New("empty command")
	}

	parts := strings.Fields(line)
	commandStr := strings.ToUpper(parts[0])

	var command CommandType

	switch commandStr {
	case "GET":
		command = GetCommand

		if len(parts) != 2 {
			return "", uuid.Nil, nil, errors.New("GET command requires exactly one argument (key)")
		}
	case "SET":
		command = SetCommand

		if len(parts) < 3 {
			return "", uuid.Nil, nil, errors.New("SET command requires at least two arguments (key and value)")
		}
	case "DELETE":
		command = DeleteCommand

		if len(parts) != 2 {
			return "", uuid.Nil, nil, errors.New("DELETE command requires exactly one argument (key)")
		}
	default:
		return "", uuid.Nil, nil, fmt.Errorf("unknown command: %s", commandStr)
	}

	keyStr := parts[1]

	key, err := uuid.Parse(keyStr)
	if err != nil {
		return "", uuid.Nil, nil, fmt.Errorf("invalid UUID: %w", err)
	}

	var value *string

	if command == SetCommand {
		// Reconstruct value from parts after the key.
		// This handles values with spaces.
		valueStr := strings.Join(parts[2:], " ")
		value = &valueStr
	}

	return command, key, value, nil
}

func getNewConn(ctx context.Context, nodes []warehouse.NodeInfo) (*grpc.ClientConn, []warehouse.NodeInfo) {
	var (
		conn     *grpc.ClientConn
		newNodes []warehouse.NodeInfo
		err      error
	)

	for _, node := range nodes {
		conn, err = grpc.NewClient(node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}

		mesh, err := warehousepb.NewWarehouseClient(conn).GetKnownNodes(ctx, &emptypb.Empty{})
		if err != nil {
			continue
		}

		newNodes, err = convertMesh(mesh.GetNodes())
		if err != nil {
			continue
		}

		log.Println("Switched to a new node:", node.UUID.String())

		break
	}

	return conn, newNodes
}
