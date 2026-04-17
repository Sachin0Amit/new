package etl

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/pkg/logger"
)

// Job represents a single item moving through the cognitive pipeline.
type Job struct {
	ID      uuid.UUID
	Payload interface{}
	Error   error
}

// Stage represents a transformative operation on a Job.
type Stage func(ctx context.Context, input interface{}) (interface{}, error)

// Pipeline orchestrates the flow of data through multiple cognitive stages in parallel.
type Pipeline struct {
	stages []Stage
	concurrency int
	logger logger.Logger
}

// NewPipeline creates a new ETL pipeline with specified concurrency.
func NewPipeline(concurrency int) *Pipeline {
	return &Pipeline{
		concurrency: concurrency,
		logger:      logger.New(),
	}
}

// AddStage appends a transformative step to the pipeline.
func (p *Pipeline) AddStage(s Stage) {
	p.stages = append(p.stages, s)
}

// Process runs a slice of input data through the pipeline using parallel workers.
func (p *Pipeline) Process(ctx context.Context, source []interface{}) <-chan Job {
	out := make(chan Job, len(source))
	
	var wg sync.WaitGroup
	jobs := make(chan interface{}, len(source))

	// Initial producer
	go func() {
		for _, item := range source {
			jobs <- item
		}
		close(jobs)
	}()

	// Worker pool
	wg.Add(p.concurrency)
	for i := 0; i < p.concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			for item := range jobs {
				current := item
				var err error
				
				jobID := uuid.New()
				p.logger.Debug("Processing Job", logger.String("job_id", jobID.String()), logger.Int("worker", id))

				for _, stage := range p.stages {
					current, err = stage(ctx, current)
					if err != nil {
						out <- Job{ID: jobID, Error: err}
						goto nextJob
					}
				}
				
				out <- Job{ID: jobID, Payload: current}
			nextJob:
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
