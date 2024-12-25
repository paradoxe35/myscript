package utils

import (
	"log"
	"os"
	"sync"
)

type ProcessQueue struct {
	processCounter uint
	processes      map[BookID]*process
	processing     bool
	lock           sync.Mutex
	executionId    BookID

	logger *log.Logger
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
		logger:         log.New(os.Stdout, name, log.LstdFlags),
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
		log.Printf("ProcessQueue: process %d not found", id)
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
		// If process not found, it means that the worker was stopped
		// reset the processing flag
		pq.processes = make(map[BookID]*process)
		pq.processCounter = 0
		pq.executionId = 0
		pq.processing = false
		return
	}

	// Wait for the process to be ready
	<-proc.ready

	pq.logger.Printf("ProcessQueue: executing process %d\n", pq.executionId)

	proc.callback()

	delete(pq.processes, pq.executionId)

	// Recursively call the worker
	go pq.worker()
}
