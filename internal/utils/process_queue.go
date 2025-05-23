// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package utils

import (
	"log/slog"
	"sync"
)

type ProcessQueue struct {
	processCounter uint
	processes      map[BookID]*process
	processing     bool
	lock           sync.Mutex
	executionId    BookID
	name           string
}

type process struct {
	ready    chan bool
	callback func()
}

type BookID uint

func NewProcessQueue(name string) *ProcessQueue {
	return &ProcessQueue{
		processCounter: 0,
		processes:      make(map[BookID]*process),
		processing:     false,
		lock:           sync.Mutex{},
		name:           name,
	}
}

func (pq *ProcessQueue) Book() BookID {
	pq.processCounter++
	newId := BookID(pq.processCounter)

	pq.processes[newId] = &process{
		ready:    make(chan bool),
		callback: nil,
	}

	if !pq.processing {
		go pq.worker()
	}

	return newId
}

func (pq *ProcessQueue) Add(id BookID, callback func()) {
	if _, ok := pq.processes[id]; !ok {
		slog.Debug("ProcessQueue: process not found", "name", pq.name, "id", uint(id))
		return
	}

	pq.processes[id].callback = callback

	// To avoid deadlock
	go func() {
		pq.processes[id].ready <- true
	}()
}

func (pq *ProcessQueue) worker() {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	pq.processing = true
	pq.executionId++

	proc, ok := pq.processes[pq.executionId]
	if !ok {
		// If process not found, it means we reached the end of the queue
		pq.processes = make(map[BookID]*process)
		pq.processCounter = 0
		pq.executionId = 0
		pq.processing = false
		return
	}

	// Wait for the process to be ready
	<-proc.ready

	slog.Debug("ProcessQueue: executing process", "name", pq.name, "id", uint(pq.executionId))

	if proc.callback != nil {
		proc.callback()
	}

	delete(pq.processes, pq.executionId)

	// Recursively call the worker
	go pq.worker()
}
