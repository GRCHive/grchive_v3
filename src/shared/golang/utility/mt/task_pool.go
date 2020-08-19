package mt

import (
	"sync"
)

type Job interface {
	Do() error
}

type TaskPool struct {
	ConcurrentJobs int

	jobs chan Job
}

func NewTaskPool(concurrentJobs int, totalTasks int) *TaskPool {
	return &TaskPool{
		ConcurrentJobs: concurrentJobs,
		jobs:           make(chan Job, totalTasks),
	}
}

func (t *TaskPool) AddJob(j Job) {
	t.jobs <- j
}

func (t *TaskPool) SyncExecute() error {
	close(t.jobs)

	wg := sync.WaitGroup{}

	allErrs := make(chan error, cap(t.jobs))

	for i := 1; i <= t.ConcurrentJobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range t.jobs {
				err := j.Do()
				if err != nil {
					allErrs <- err
				}
			}
		}()
	}

	wg.Wait()
	close(allErrs)

	for err := range allErrs {
		return err
	}
	return nil
}
