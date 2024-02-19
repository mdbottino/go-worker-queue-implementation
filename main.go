package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Job represents the data we're going to work on
type Job struct {
	filename string
	wg       *sync.WaitGroup
}

// A buffered channel that we can send work requests on
var JobQueue chan Job = make(chan Job)

// Worker represents the worker that executes the job
type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool)}
}

// Starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w Worker) Start() {
	go func() {
		for {
			// Register the current worker into the worker queue
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// We have received a work request
				read(job.filename, job.wg)

			case <-w.quit:
				// We have received a signal to stop
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

// Dispatcher is the orchestrator that is responsible for starting and stopping workers.
// It is also the one that sends the job to the worker
type Dispatcher struct {
	workerPool chan chan Job
	maxWorkers int
	workers    []Worker
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{workerPool: pool, maxWorkers: maxWorkers}
}

// Start all workers and start the main dispatch loop
func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.workerPool)
		worker.Start()
		d.workers = append(d.workers, worker)
	}

	go d.dispatch()
}

// Stop all workers
func (d *Dispatcher) Stop() {
	for _, worker := range d.workers {
		worker.Stop()
	}
}

// Listen for a new job and place it into the first available worker queue
func (d *Dispatcher) dispatch() {
	for job := range JobQueue {
		// A job request has been received
		go func(job Job) {
			// Get a worker job channel that is available. this will block
			// until a worker is idle
			jobChannel := <-d.workerPool

			// Dispatch the job to the worker job channel
			jobChannel <- job
		}(job)
	}
}

func read(name string, wg *sync.WaitGroup) {
	defer wg.Done()

	data, err := os.ReadFile(name)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	if string(data) != "Hello, world" {
		fmt.Println("File content mismatch: got", string(data), ", want 'Hello, world'")
		return
	}
}

const filename string = "./text.txt"
const rounds int = 10000
const workerCount int = 5

func main() {
	startTime := time.Now()

	// create a WaitGroup so we can make sure we exit once all the work is done
	wg := new(sync.WaitGroup)
	wg.Add(rounds)

	dispatcher := NewDispatcher(workerCount)
	dispatcher.Run()

	for i := 0; i < rounds; i++ {

		// Creating the job with the appropriate payload
		work := Job{filename, wg}

		// Push the work onto the queue.
		JobQueue <- work
	}

	wg.Wait()
	dispatcher.Stop()

	fmt.Printf("Execution time: %v\n", time.Since(startTime))
}
