package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/api/warehousepb"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/warehouse"
)

type Client struct {
	conn     *grpc.ClientConn
	addr     string
	nodeUUID uuid.UUID
}

var (
	ErrConfMismatch = errors.New("node config didn't match")
)

func NewClient(addr string, nodeUUID uuid.UUID) *Client {
	return &Client{
		conn:     nil,
		addr:     addr,
		nodeUUID: nodeUUID,
	}
}

func (client *Client) GetUUID() uuid.UUID {
	return client.nodeUUID
}

func (client *Client) GetAddress() string {
	return client.addr
}

func (client *Client) SendPulse(ctx context.Context, from warehouse.NodeInfo) (warehouse.FullNodeInfo, error) {
	err := client.connect()
	if err != nil {
		return warehouse.FullNodeInfo{}, fmt.Errorf("connect: %w", err)
	}

	resp, err := warehousepb.NewWarehouseClient(client.conn).PulseNodeToNode(ctx, &warehousepb.PulseRequest{
		Requester: &warehousepb.Node{
			NodeUuid: from.UUID.String(),
			Address:  from.Address,
		},
	})
	if err != nil {
		return warehouse.FullNodeInfo{}, fmt.Errorf("send pulse: %w", err)
	}

	nodeID, err := uuid.Parse(resp.GetRespondent().GetNodeUuid())
	if err != nil {
		return warehouse.FullNodeInfo{}, fmt.Errorf("parse node uuid: %w", err)
	}

	pulsePeriod, err := time.ParseDuration(resp.GetRespondentPulsePeriod())
	if err != nil {
		return warehouse.FullNodeInfo{}, fmt.Errorf("parse pulse perion: %w", err)
	}

	return warehouse.FullNodeInfo{
		PulseInterval: pulsePeriod,
		NodeInfo: warehouse.NodeInfo{
			Address: resp.GetRespondent().GetAddress(),
			UUID:    nodeID,
		},
	}, nil
}

func (client *Client) GetKnownNodes(ctx context.Context) ([]warehouse.NodeConn, error) {
	err := client.connect()
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	resp, err := warehousepb.NewWarehouseClient(client.conn).GetKnownNodes(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("get known hosts: %w", err)
	}

	var (
		meshNodeID uuid.UUID
		result     = make([]warehouse.NodeConn, 0, len(resp.GetNodes()))
	)

	for _, node := range resp.GetNodes() {
		meshNodeID, err = uuid.Parse(node.GetNodeUuid())
		if err != nil {
			return nil, fmt.Errorf("parse node uuid: %w", err)
		}

		result = append(result, NewClient(node.GetAddress(), meshNodeID))
	}

	return result, nil
}

func (client *Client) connect() error {
	if client.conn != nil {
		return nil
	}

	conn, err := grpc.NewClient(client.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	client.conn = conn

	return nil
}

func (client *Client) FixMesh(ctx context.Context, replicaFactor int) {
	err := client.connect()
	if err != nil {
		return
	}

	_, err = warehousepb.NewWarehouseClient(client.conn).FixMesh(ctx, &warehousepb.FixMeshRequest{
		ReplicaFactor: int32(replicaFactor),
	})
	if err != nil {
		return
	}
}
func (client *Client) StorageDelete(ctx context.Context, key uuid.UUID) (bool, error) {
	err := client.connect()
	if err != nil {
		return false, fmt.Errorf("connect: %w", err)
	}

	resp, err := warehousepb.NewWarehouseClient(client.conn).StorageDelete(ctx, &warehousepb.StorageDeleteRequest{
		Key: key.String(),
	})
	if err != nil {
		return false, fmt.Errorf("client storage delete: %w", err)
	}

	return resp.GetFound(), nil
}

func (client *Client) StorageLoad(ctx context.Context, key uuid.UUID) ([]byte, bool, error) {
	err := client.connect()
	if err != nil {
		return nil, false, fmt.Errorf("connect: %w", err)
	}

	resp, err := warehousepb.NewWarehouseClient(client.conn).StorageLoad(ctx, &warehousepb.StorageLoadRequest{
		Key: key.String(),
	})
	if err != nil {
		return nil, false, fmt.Errorf("client storage load: %w", err)
	}

	data := []byte(nil)
	if resp.GetFound() && resp.GetValuePresent() != nil {
		data = []byte(resp.GetValue())
	}

	return data, resp.GetFound(), nil
}

func (client *Client) StorageStore(ctx context.Context, key uuid.UUID, value []byte) error {
	err := client.connect()
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	_, err = warehousepb.NewWarehouseClient(client.conn).StorageStore(ctx, &warehousepb.StorageStoreRequest{
		Key:   key.String(),
		Value: string(value),
	})
	if err != nil {
		return fmt.Errorf("client storage store: %w", err)
	}

	return nil
}
