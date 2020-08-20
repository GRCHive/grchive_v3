package mt

import (
	"sync"
)

type Job interface {
	Do() error
}

type TaskPool struct {
	ConcurrentJobs int

	jobs []Job
}

func NewTaskPool(concurrentJobs int) *TaskPool {
	return &TaskPool{
		ConcurrentJobs: concurrentJobs,
		jobs:           []Job{},
	}
}

func (t *TaskPool) AddJob(j Job) {
	t.jobs = append(t.jobs, j)
}

func (t *TaskPool) SyncExecute() error {
	wg := sync.WaitGroup{}

	jobChan := make(chan Job)
	allErrs := make(chan error, cap(t.jobs))

	for i := 1; i <= t.ConcurrentJobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobChan {
				err := j.Do()
				if err != nil {
					allErrs <- err
				}
			}
		}()
	}

	for _, j := range t.jobs {
		jobChan <- j
	}
	close(jobChan)

	wg.Wait()
	close(allErrs)

	for err := range allErrs {
		return err
	}
	return nil
}
