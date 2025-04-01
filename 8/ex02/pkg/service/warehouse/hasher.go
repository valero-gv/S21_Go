package warehouse

import (
	"bytes"
	"hash/fnv"
	"slices"

	"github.com/google/uuid"
)

type OnlineNode struct {
	NodeInfo
	Online bool
}

type Hasher struct {
	knownNodes        map[uuid.UUID]int
	knownNodesAddress []OnlineNode
	replicaFactor     int
}

func NewHasher(initNodes []NodeInfo, replicaFactor int) *Hasher {
	result := &Hasher{
		knownNodes:        make(map[uuid.UUID]int, len(initNodes)),
		knownNodesAddress: make([]OnlineNode, 0, len(initNodes)),
		replicaFactor:     replicaFactor,
	}

	if result.replicaFactor > len(initNodes) {
		result.replicaFactor = len(initNodes)
	}

	slices.SortFunc(initNodes, func(a, b NodeInfo) int {
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

func (h Hasher) UpdateOnlines(nodes []NodeInfo) error {
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

func (h Hasher) KeyToNodes(key uuid.UUID) []NodeInfo {
	nodeIndices := make([]NodeInfo, 0, h.replicaFactor)
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
