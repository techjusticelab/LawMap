package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"lawmap/internal/etl"
	"lawmap/internal/pkg/log"
)

// Scheduler manages periodic execution of ETL jobs.
type Scheduler struct {
	registry      *etl.Registry
	jobs          map[string]*etl.Job
	mu            sync.RWMutex
	maxConcurrent int
	running       int
	logger        *log.Logger
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// Config holds scheduler configuration.
type Config struct {
	MaxConcurrent    int
	RetryAttempts    int
	RetryDelayMinutes int
}

// New creates a new scheduler with the given configuration.
func New(registry *etl.Registry, cfg Config) *Scheduler {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 3
	}
	return &Scheduler{
		registry:      registry,
		jobs:          make(map[string]*etl.Job),
		maxConcurrent: cfg.MaxConcurrent,
		logger:        log.Default().WithField("component", "scheduler"),
		stopChan:      make(chan struct{}),
	}
}

// RunOnce executes a single ETL job immediately.
func (s *Scheduler) RunOnce(ctx context.Context, pipelineName string) (*etl.Result, error) {
	pipeline, ok := s.registry.Get(pipelineName)
	if !ok {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineName)
	}

	jobID := fmt.Sprintf("%s-%d", pipelineName, time.Now().Unix())
	job := &etl.Job{
		ID:         jobID,
		SourceName: pipelineName,
		Status:     etl.StatusRunning,
		StartedAt:  time.Now(),
	}

	s.mu.Lock()
	s.jobs[jobID] = job
	s.mu.Unlock()

	s.logger.Info("Starting ETL job", map[string]any{"job_id": jobID, "pipeline": pipelineName})

	result, err := pipeline.Run(ctx)
	job.FinishedAt = time.Now()
	job.Result = result

	if err != nil {
		job.Status = etl.StatusFailed
		job.Error = err
		s.logger.Error("ETL job failed", map[string]any{"job_id": jobID, "error": err.Error()})
		return result, err
	}

	job.Status = etl.StatusCompleted
	s.logger.Info("ETL job completed", map[string]any{
		"job_id":         jobID,
		"nodes_created":  result.NodesCreated,
		"edges_created":  result.EdgesCreated,
		"duration_secs":  result.EndTime.Sub(result.StartTime).Seconds(),
	})

	return result, nil
}

// RunAsync executes an ETL job in the background.
func (s *Scheduler) RunAsync(pipelineName string) (string, error) {
	s.mu.Lock()
	if s.running >= s.maxConcurrent {
		s.mu.Unlock()
		return "", fmt.Errorf("max concurrent jobs reached (%d)", s.maxConcurrent)
	}
	s.running++
	s.mu.Unlock()

	jobID := fmt.Sprintf("%s-%d", pipelineName, time.Now().Unix())
	job := &etl.Job{
		ID:         jobID,
		SourceName: pipelineName,
		Status:     etl.StatusPending,
	}

	s.mu.Lock()
	s.jobs[jobID] = job
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() {
			s.mu.Lock()
			s.running--
			s.mu.Unlock()
		}()

		ctx := context.Background()
		_, _ = s.RunOnce(ctx, pipelineName)
	}()

	return jobID, nil
}

// GetJob retrieves job status by ID.
func (s *Scheduler) GetJob(jobID string) (*etl.Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[jobID]
	return job, ok
}

// ListJobs returns all jobs.
func (s *Scheduler) ListJobs() []*etl.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	jobs := make([]*etl.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	s.logger.Info("Scheduler stopped")
}
