package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"sync"
	"sync/atomic"
)

type routinePool struct {
	wg        sync.WaitGroup
	taskQueue chan common.Hash

	emptyContract   atomic.Uint64
	arrCount        atomic.Uint64
	varCount        atomic.Uint64
	arrItemCount    atomic.Uint64
	arrItemCountMap map[int]uint64
	mapLock         *sync.Mutex
}

func NewRoutinePool(size int64, snap snapshot.Snapshot, keySet []common.Hash) *routinePool {
	pool := &routinePool{
		taskQueue:       make(chan common.Hash),
		arrCount:        atomic.Uint64{},
		varCount:        atomic.Uint64{},
		arrItemCount:    atomic.Uint64{},
		arrItemCountMap: make(map[int]uint64),
		mapLock:         &sync.Mutex{},
	}

	for i := int64(0); i < size; i++ {
		go pool.worker(snap, keySet)
	}
	return pool
}

func (pool *routinePool) AddTask(addrHash common.Hash) {
	pool.wg.Add(1)
	pool.taskQueue <- addrHash
}

func (pool *routinePool) worker(snap snapshot.Snapshot, keySet []common.Hash) {
	for addrHash := range pool.taskQueue {
		arrCount, arrItemCount, varCount := traversalContract(snap, addrHash, keySet, pool.mapLock, pool.arrItemCountMap)
		pool.arrCount.Add(uint64(arrCount))
		pool.arrItemCount.Add(uint64(arrItemCount))
		pool.varCount.Add(uint64(varCount))

		if arrCount == 0 && varCount == 0 {
			pool.emptyContract.Add(1)
		}
		pool.wg.Done()
	}
}

// Wait
func (pool *routinePool) Wait() {
	pool.wg.Wait()
}
