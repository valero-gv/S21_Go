package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/api/warehousepb"
	"repos.21-school.ru/students/Go_Day09.ID_376232/dougiela/Go_Day09-1/src/ex02/pkg/service/warehouse"
)

type WarehouseServer struct {
	me *warehouse.Node
	warehousepb.UnimplementedWarehouseServer
}

func NewServer(node *warehouse.Node) *WarehouseServer {
	return &WarehouseServer{
		me: node,
	}
}

func (server WarehouseServer) GetKnownNodes(_ context.Context, _ *emptypb.Empty,
) (*warehousepb.GetKnownNodesResponse, error) {
	nodes, _ := server.me.GetKnownNodes()
	result := make([]*warehousepb.Node, 0, len(nodes))

	for _, node := range nodes {
		result = append(result, &warehousepb.Node{
			NodeUuid: node.UUID.String(),
			Address:  node.Address,
		})
	}

	result = append(result, &warehousepb.Node{
		NodeUuid: server.me.GetNodeUUID().String(),
		Address:  server.me.GetAddress(),
	})

	return &warehousepb.GetKnownNodesResponse{
		Nodes: result,
		Pulse: durationpb.New(server.me.GetPulseInterval()),
	}, nil
}

func (server WarehouseServer) PulseNodeToNode(_ context.Context, request *warehousepb.PulseRequest,
) (*warehousepb.PulseResponse, error) {
	requesterUUID, err := uuid.Parse(request.GetRequester().GetNodeUuid())
	if err != nil {
		return nil, fmt.Errorf("parse requester uuid: %w", err)
	}

	server.me.IncomingPulse(NewClient(request.GetRequester().GetAddress(), requesterUUID))

	return &warehousepb.PulseResponse{
		Respondent: &warehousepb.Node{
			NodeUuid: server.me.GetNodeUUID().String(),
			Address:  server.me.GetAddress(),
		},
		RespondentPulsePeriod: server.me.GetPulseInterval().String(),
	}, nil
}

func (server WarehouseServer) StorageStore(_ context.Context, request *warehousepb.StorageStoreRequest,
) (*emptypb.Empty, error) {
	key, err := uuid.Parse(request.GetKey())
	if err != nil {
		return nil, fmt.Errorf("parse key uuid: %w", err)
	}

	server.me.StorageStore(key, []byte(request.GetValue()))

	return &emptypb.Empty{}, nil
}

func (server WarehouseServer) StorageLoad(_ context.Context, request *warehousepb.StorageLoadRequest,
) (*warehousepb.StorageLoadResponse, error) {
	key, err := uuid.Parse(request.GetKey())
	if err != nil {
		return nil, fmt.Errorf("parse key uuid: %w", err)
	}

	val, found := server.me.StorageLoad(key)
	result := &warehousepb.StorageLoadResponse{
		Found:        found,
		ValuePresent: nil,
	}

	if found {
		result.ValuePresent = &warehousepb.StorageLoadResponse_Value{Value: string(val)}
	}

	return result, nil
}

func (server WarehouseServer) StorageDelete(_ context.Context, request *warehousepb.StorageDeleteRequest,
) (*warehousepb.StorageDeleteResponse, error) {
	key, err := uuid.Parse(request.GetKey())
	if err != nil {
		return nil, fmt.Errorf("parse key uuid: %w", err)
	}

	found := server.me.StorageDelete(key)

	return &warehousepb.StorageDeleteResponse{
		Found: found,
	}, nil
}

func (server WarehouseServer) ClientGetNodes(_ context.Context, _ *emptypb.Empty,
) (*warehousepb.ClientGetNodesResponse, error) {
	nodes, boss := server.me.GetKnownNodes()
	result := make([]*warehousepb.Node, 0, len(nodes))

	for _, node := range nodes {
		result = append(result, &warehousepb.Node{
			NodeUuid: node.UUID.String(),
			Address:  node.Address,
		})
	}

	result = append(result, &warehousepb.Node{
		NodeUuid: server.me.GetNodeUUID().String(),
		Address:  server.me.GetAddress(),
	})

	return &warehousepb.ClientGetNodesResponse{
		Nodes: result,
		Leader: &warehousepb.Node{
			NodeUuid: boss.UUID.String(),
			Address:  boss.Address,
		},
	}, nil
}

func (server WarehouseServer) ClientStorageStore(ctx context.Context, request *warehousepb.StorageStoreRequest,
) (*warehousepb.ClientStorageStoreResponse, error) {
	key, err := uuid.Parse(request.GetKey())
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	replicas, err := server.me.ClientStorageStore(ctx, key, []byte(request.GetValue()))
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return &warehousepb.ClientStorageStoreResponse{
		StoredReplicas: int32(replicas),
	}, nil
}

func (server WarehouseServer) ClientStorageLoad(ctx context.Context, request *warehousepb.StorageLoadRequest,
) (*warehousepb.StorageLoadResponse, error) {
	key, err := uuid.Parse(request.GetKey())
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	value, found, err := server.me.ClientStorageLoad(ctx, key)
	if errors.Is(err, warehouse.ErrNotFound) {
		return &warehousepb.StorageLoadResponse{Found: false, ValuePresent: nil}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	result := &warehousepb.StorageLoadResponse{
		Found:        found,
		ValuePresent: nil,
	}
	if found {
		result.ValuePresent = &warehousepb.StorageLoadResponse_Value{Value: string(value)}
	}

	return result, nil
}

func (server WarehouseServer) ClientStorageDelete(ctx context.Context, request *warehousepb.StorageDeleteRequest,
) (*warehousepb.ClientStorageDeleteResponse, error) {
	key, err := uuid.Parse(request.GetKey())
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	replicas, err := server.me.ClientStorageDelete(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("delete: %w", err)
	}

	return &warehousepb.ClientStorageDeleteResponse{
		DeletedReplicas: int32(replicas),
	}, nil
}

func (server WarehouseServer) FixMesh(ctx context.Context, request *warehousepb.FixMeshRequest,
) (*emptypb.Empty, error) {
	server.me.FixMesh(ctx, int(request.GetReplicaFactor()))

	return &emptypb.Empty{}, nil
}
