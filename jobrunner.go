package jobrunner

import (
	"bytes"
	"log"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type IJob interface {
	Run() error
}

type Job struct {
	inner  IJob
	status uint32

	running sync.Mutex

	// web
	Name       string
	Status     string
	Latency    string
	JobEntryID int
}

const UNNAMED = "(unnamed)"

func New(name string, job IJob) *Job {
	jobName := name
	if jobName == "" {
		jobName := reflect.TypeOf(job).Name()
		if jobName == "Func" {
			jobName = UNNAMED
		}
	}

	return &Job{
		Name:  name,
		inner: job,
	}
}

func (j *Job) StatusUpdate() string {
	if atomic.LoadUint32(&j.status) > 0 {
		j.Status = "RUNNING"
		return j.Status
	}
	j.Status = "IDLE"
	return j.Status

}

func (j *Job) Run() {
	start := time.Now()
	// If the job panics, just print a stack trace.
	// Don't let the whole process die.
	defer func() {
		if err := recover(); err != nil {
			var buf bytes.Buffer
			logger := log.New(&buf, "JobRunner Log: ", log.Lshortfile)
			logger.Panic(err, "\n", string(debug.Stack()))
		}
	}()

	if !selfConcurrent {
		j.running.Lock()
		defer j.running.Unlock()
	}

	if workPermits != nil {
		workPermits <- struct{}{}
		defer func() { <-workPermits }()
	}

	atomic.StoreUint32(&j.status, 1)
	j.StatusUpdate()

	defer j.StatusUpdate()
	defer atomic.StoreUint32(&j.status, 0)

	j.inner.Run()

	end := time.Now()
	j.Latency = end.Sub(start).String()

}
