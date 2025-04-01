package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/api/warehousepb"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/warehouse"
)

type appConf struct {
	initialNodeAddress string
	initialNodeUUID    uuid.UUID
	replicaFactor      int
	pulseInterval      time.Duration
}

type CommandType string

const (
	GetCommand    CommandType = "GET"
	SetCommand    CommandType = "SET"
	DeleteCommand CommandType = "DELETE"
	FixCommand    CommandType = "FIX"
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
		log.Println("Err Read conf: ", err.Error())

		return
	}

	err = startREPL(appCtx, conf)
	if err != nil {
		log.Printf("Err Run repl: %s\n", err)

		return
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

	leader, nodes, err := getNewLeader(ctx, LeaderConn{
		conn: conn,
		id:   conf.initialNodeUUID.String(),
	}, nil)
	if err != nil {
		return fmt.Errorf("get initial leader: %w", err)
	}

	err = repl(ctx, conf.replicaFactor, conf.pulseInterval, leader, nodes)
	if err != nil {
		return fmt.Errorf("do repl: %w", err)
	}

	return nil
}

func repl(ctx context.Context, replicaFactor int, pulseInterval time.Duration, leader LeaderConn,
	nodes []warehouse.NodeInfo,
) error {
	inpChan := make(chan string)
	releaderTicker := time.NewTicker(pulseInterval)
	defer releaderTicker.Stop()

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
			err := handleCommand(ctx, stdin, leader.conn, replicaFactor)
			if err != nil {
				log.Println("Err Handle command:", err)
			}
		case <-releaderTicker.C:
			var err error

			leader, nodes, err = getNewLeader(ctx, leader, nodes)
			if err != nil {
				return fmt.Errorf("get new leader: %w", err)
			}
			log.Println("Leader:", leader.id)

			if len(nodes) < replicaFactor {
				log.Println("Mesh size is lower than the replica factor -", len(nodes))
			}
		}
	}

	return nil
}

func handleCommand(ctx context.Context, command string, leader *grpc.ClientConn, replicaFactor int) error {
	commandType, key, value, err := parseCommand(command)
	if err != nil {
		return fmt.Errorf("parse command: %w", err)
	}

	switch commandType {
	case FixCommand:
		_, err := warehousepb.NewWarehouseClient(leader).FixMesh(ctx, &warehousepb.FixMeshRequest{
			ReplicaFactor: int32(replicaFactor),
		})
		if err != nil {
			return fmt.Errorf("fix mesh: %w", err)
		}

	case SetCommand:
		resp, err := warehousepb.NewWarehouseClient(leader).ClientStorageStore(ctx, &warehousepb.StorageStoreRequest{
			Key:   key.String(),
			Value: value,
		})
		if err != nil {
			return fmt.Errorf("client store: %w", err)
		}

		log.Printf("Stored %d replicas\n", resp.GetStoredReplicas())

	case GetCommand:
		resp, err := warehousepb.NewWarehouseClient(leader).ClientStorageLoad(ctx, &warehousepb.StorageLoadRequest{
			Key: key.String(),
		})
		if err != nil {
			return fmt.Errorf("client load: %w", err)
		}

		if !resp.GetFound() {
			log.Println("Not found")
		} else {
			log.Println(resp.GetValue())
		}

	case DeleteCommand:
		resp, err := warehousepb.NewWarehouseClient(leader).ClientStorageDelete(ctx, &warehousepb.StorageDeleteRequest{
			Key: key.String(),
		})
		if err != nil {
			return fmt.Errorf("client store: %w", err)
		}

		log.Printf("Deleted %d replicas\n", resp.GetDeletedReplicas())
	}

	return nil
}

func parseCommand(line string) (CommandType, uuid.UUID, string, error) {
	line = strings.TrimSpace(line)

	if line == "" {
		return "", uuid.Nil, "", errors.New("empty command")
	}

	parts := strings.Fields(line)
	commandStr := strings.ToUpper(parts[0])

	var command CommandType

	switch commandStr {
	case "GET":
		command = GetCommand

		if len(parts) != 2 {
			return "", uuid.Nil, "", errors.New("GET command requires exactly one argument (key)")
		}
	case "SET":
		command = SetCommand

		if len(parts) < 3 {
			return "", uuid.Nil, "", errors.New("SET command requires at least two arguments (key and value)")
		}
	case "DELETE":
		command = DeleteCommand

		if len(parts) != 2 {
			return "", uuid.Nil, "", errors.New("DELETE command requires exactly one argument (key)")
		}
	case "FIX":
		command = FixCommand

		return command, uuid.Nil, "", nil
	default:
		return "", uuid.Nil, "", fmt.Errorf("unknown command: %s", commandStr)
	}

	keyStr := parts[1]

	key, err := uuid.Parse(keyStr)
	if err != nil {
		return "", uuid.Nil, "", fmt.Errorf("invalid UUID: %w", err)
	}

	var value string

	if command == SetCommand {
		value = strings.Join(parts[2:], " ")
	}

	return command, key, value, nil
}

type LeaderConn struct {
	conn *grpc.ClientConn
	id   string
}

func getNewLeader(ctx context.Context, conn LeaderConn, known []warehouse.NodeInfo,
) (LeaderConn, []warehouse.NodeInfo, error) {
	handleNodeResp := func(conn LeaderConn, resp *warehousepb.ClientGetNodesResponse,
	) (LeaderConn, []warehouse.NodeInfo, error) {
		var err error

		if resp.GetLeader().GetNodeUuid() != conn.id {
			conn.id = resp.GetLeader().GetNodeUuid()
			conn.conn.Close()

			conn.conn, err = grpc.NewClient(resp.GetLeader().GetAddress(),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return LeaderConn{}, nil, fmt.Errorf("open conn to the new leader: %w", err)
			}
		}

		mesh, err := convertMesh(resp.GetNodes())
		if err != nil {
			return LeaderConn{}, nil, fmt.Errorf("convert nodes: %w", err)
		}

		return conn, mesh, nil
	}

	resp, err := warehousepb.NewWarehouseClient(conn.conn).ClientGetNodes(ctx, &emptypb.Empty{})
	if err == nil {
		return handleNodeResp(conn, resp)
	}

	var nextConn *grpc.ClientConn

	for _, node := range known {
		nextConn, err = grpc.NewClient(node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}

		resp, err = warehousepb.NewWarehouseClient(nextConn).ClientGetNodes(ctx, &emptypb.Empty{})
		if err != nil {
			continue
		}

		conn, known, err = handleNodeResp(conn, resp)
		if err == nil {
			return conn, known, nil
		}
	}

	return LeaderConn{}, nil, ErrNoMoreNodes
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
