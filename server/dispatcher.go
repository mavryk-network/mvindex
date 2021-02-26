// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package server

// A buffered channel that we can send work requests on.
var jobQueue chan *ApiContext

type Worker struct {
	WorkerPool chan chan *ApiContext
	JobChannel chan *ApiContext
	quit       chan bool
}

func NewWorker(workerPool chan chan *ApiContext) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan *ApiContext),
		quit:       make(chan bool)}
}

func (w Worker) Start() {
	go func() {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel

			// wait for next job or shutdown
			select {
			case req := <-w.JobChannel:
				select {
				// check if the context is expired
				case <-req.Context.Done():
					err := req.Context.Err()
					req.handleError(err)
					req.sendResponse()
				default:
					req.serve()
					req.sendResponse()
				}
				req.done <- nil

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	pool       chan chan *ApiContext
	maxWorkers int
	maxQueue   int
}

func NewDispatcher(maxWorkers int, maxQueue int) *Dispatcher {
	jobQueue = make(chan *ApiContext, maxQueue)
	pool := make(chan chan *ApiContext, maxWorkers)
	return &Dispatcher{pool: pool, maxWorkers: maxWorkers, maxQueue: maxQueue}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.pool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case api := <-jobQueue:
			// try to obtain a worker job channel that is available.
			// will block until a worker is idle or pool is closed
			workerChannel := <-d.pool

			// dispatch the job to the worker job channel
			if workerChannel != nil {
				workerChannel <- api
			}
		}
	}
}
