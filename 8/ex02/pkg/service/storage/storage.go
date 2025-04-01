package storage

import (
	"sync"

	"github.com/google/uuid"
)

type InMemo struct {
	storage   map[uuid.UUID][]byte
	storageMU sync.RWMutex
}

func NewMemoStorage() *InMemo {
	return &InMemo{
		storage:   make(map[uuid.UUID][]byte),
		storageMU: sync.RWMutex{},
	}
}

func (memo *InMemo) GetSize() int {
	memo.storageMU.RLock()
	defer memo.storageMU.RUnlock()

	return len(memo.storage)
}

func (memo *InMemo) Load(key uuid.UUID) ([]byte, bool) {
	memo.storageMU.RLock()
	defer memo.storageMU.RUnlock()

	val, ok := memo.storage[key]

	return val, ok
}

func (memo *InMemo) Store(key uuid.UUID, value []byte) {
	memo.storageMU.Lock()
	defer memo.storageMU.Unlock()

	memo.storage[key] = value
}

func (memo *InMemo) Delete(key uuid.UUID) bool {
	memo.storageMU.Lock()
	defer memo.storageMU.Unlock()

	_, found := memo.storage[key]

	delete(memo.storage, key)

	return found
}
