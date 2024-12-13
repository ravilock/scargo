package worker

import (
	"fmt"
	"time"
)

type Result[FinalResult any] struct {
	Output   FinalResult
	Err      error
	WorkerID int
}

type WorkerPool[Task func(Args) (FinalResult, error), Args any, FinalResult any] struct {
	TaskQueue   chan Args
	ResultChan  chan Result[FinalResult]
	workers     []*Worker[Task, Args, FinalResult]
	WorkerCount int
}

type Worker[Task func(Args) (FinalResult, error), Args any, FinalResult any] struct {
	task       Task
	taskQueue  <-chan Args
	resultChan chan<- Result[FinalResult]
	idleSince  *time.Time
	id         int
}

func (w *Worker[Task, Args, FinalResult]) Start() {
	go func() {
		for input := range w.taskQueue {
			w.idleSince = nil
			output, err := w.task(input)
			w.resultChan <- Result[FinalResult]{WorkerID: w.id, Output: output, Err: err}
			now := time.Now()
			w.idleSince = &now
		}
	}()
	go func() {
		time.Sleep(time.Second)
		now := time.Now()
		w.idleSince = &now
	}()
}

func (wp *WorkerPool[Task, Args, FinalResult]) Start(task Task) {
	wp.workers = make([]*Worker[Task, Args, FinalResult], 0, wp.WorkerCount)
	for i := 0; i < wp.WorkerCount; i++ {
		worker := &Worker[Task, Args, FinalResult]{
			id:         i,
			taskQueue:  wp.TaskQueue,
			resultChan: wp.ResultChan,
			task:       task,
		}
		worker.Start()
		wp.workers = append(wp.workers, worker)
	}
}

func (wp *WorkerPool[Task, Args, FinalResult]) Submit(input Args) {
	fmt.Println("Submitting", input)
	wp.TaskQueue <- input
}

func (wp *WorkerPool[Task, Args, FinalResult]) Finish() {
	fmt.Println("finishing...")
}
