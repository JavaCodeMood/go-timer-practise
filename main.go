package main

import (
	"time"
	"sync"
	"fmt"
)

type Task struct {
	timer *time.Timer
	duration time.Duration
	cancelChan chan byte
	f func(*Task)
	mutex sync.Mutex
	isCancel bool
	once bool
}

func (task *Task) CancelTask() {
	task.mutex.Lock()
	defer task.mutex.Unlock()
	if !task.isCancel {
		task.isCancel = true
		close(task.cancelChan)
	}
}

func AddTask(duration time.Duration, f func(*Task)) (task *Task) {
	task = &Task{}
	task.timer = time.NewTimer(duration)
	task.cancelChan = make(chan byte)
	task.isCancel = false
	task.f = f
	task.duration = duration
	task.once = false

	task.run()
	return
}

func AddOnceTask(duration time.Duration, f func(*Task)) (task *Task) {
	task = &Task{}
	task.timer = time.NewTimer(duration)
	task.cancelChan = make(chan byte)
	task.isCancel = false
	task.f = f
	task.duration = duration
	task.once = true

	task.run()
	return
}

func (task *Task) isCanceled() bool {
	task.mutex.Lock()
	defer task.mutex.Unlock()
	return task.isCancel
}

func (task *Task) isFinished() bool {
	task.mutex.Lock()
	defer task.mutex.Unlock()
	return task.isCancel || task.once
}

func (task *Task) run() {
	go func() {
		for {
			select {
			case <-task.cancelChan:
				task.timer.Stop()
				return
			case <-task.timer.C:
				if !task.isCanceled() {
					task.f(task)
					if !task.isFinished() {
						task.timer.Reset(task.duration)
					} else {
						return
					}
				}
			}
		}
	}()
}

func main() {
	var counter int = 0

	AddTask(time.Duration(time.Second * 1), func(task *Task) {
		fmt.Println("task run")
		counter++
		if counter > 10 {
			task.CancelTask()
			fmt.Println("task canceled")
		}
	})

	AddOnceTask(time.Duration(time.Second * 0), func(task *Task) {
		fmt.Println("task run once")
	})

	for {
		time.Sleep(time.Second * 1)
	}
}
