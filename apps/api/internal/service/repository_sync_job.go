package service

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)



type RepositorySyncRequest struct {
	JobID        string
	RepositoryID string
}

type RepositorySyncQueue interface {
	Enqueue(
		ctx context.Context,
		request RepositorySyncRequest,
	) (*domain.RepositorySyncJob, error)
}

type RepositorySyncMetrics struct {
	Queued    atomic.Uint64
	Retried   atomic.Uint64
	Succeeded atomic.Uint64
	Failed    atomic.Uint64
}

type RepositorySyncMetricsSnapshot struct {
	Queued    uint64
	Retried   uint64
	Succeeded uint64
	Failed    uint64
}

func (m *RepositorySyncMetrics) Snapshot() RepositorySyncMetricsSnapshot {
	return RepositorySyncMetricsSnapshot{
		Queued:    m.Queued.Load(),
		Retried:   m.Retried.Load(),
		Succeeded: m.Succeeded.Load(),
		Failed:    m.Failed.Load(),
	}
}

type RepositorySyncRetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
}

type RepositorySyncWorker struct {
	syncService RepositorySyncService
	requests    chan RepositorySyncRequest
	policy      RepositorySyncRetryPolicy
	jobs        repository.RepositorySyncJobRepository
	metrics     *RepositorySyncMetrics
}

func NewRepositorySyncWorker(
	syncService RepositorySyncService,
	queueSize int,
	policy RepositorySyncRetryPolicy,
) *RepositorySyncWorker {
	return NewRepositorySyncWorkerWithStore(
		syncService,
		queueSize,
		policy,
		nil,
	)
}

func NewRepositorySyncWorkerWithStore(
	syncService RepositorySyncService,
	queueSize int,
	policy RepositorySyncRetryPolicy,
	jobs repository.RepositorySyncJobRepository,
) *RepositorySyncWorker {
	if queueSize < 1 {
		queueSize = 1
	}

	if policy.MaxAttempts < 1 {
		policy.MaxAttempts = 1
	}

	return &RepositorySyncWorker{
		syncService: syncService,
		requests:    make(chan RepositorySyncRequest, queueSize),
		policy:      policy,
		jobs:        jobs,
		metrics:     &RepositorySyncMetrics{},
	}
}

func (w *RepositorySyncWorker) Enqueue(
	ctx context.Context,
	request RepositorySyncRequest,
) (*domain.RepositorySyncJob, error) {
	job := &domain.RepositorySyncJob{
		ID:           uuid.NewString(),
		RepositoryID: request.RepositoryID,
		Status:       domain.RepositorySyncJobStatusPending,
		Attempts:     0,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if w.jobs != nil {
		if err := w.jobs.Create(ctx, job); err != nil {
			return nil, err
		}
	}

	request.JobID = job.ID

	select {
	case w.requests <- request:
		w.metrics.Queued.Add(1)
		return job, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (w *RepositorySyncWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.drainQueuedJobs(ctx)
			return

		default:
		}

		select {
		case request := <-w.requests:
			_ = w.runWithRetry(ctx, request)

		case <-ctx.Done():
			w.drainQueuedJobs(ctx)
			return
		}
	}
}

func (w *RepositorySyncWorker) drainQueuedJobs(ctx context.Context) {
	for {
		select {
		case request := <-w.requests:
			if w.jobs != nil {
				_ = w.jobs.MarkFailed(
					ctx,
					request.JobID,
					0,
					domain.RepositorySyncJobFailureCodeWorkerShutdown,
					"worker shutting down before job was processed",
				)
			}

			w.metrics.Failed.Add(1)

		default:
			return
		}
	}
}

func (w *RepositorySyncWorker) runWithRetry(
	ctx context.Context,
	request RepositorySyncRequest,
) error {
	var err error

	for attempt := 1; attempt <= w.policy.MaxAttempts; attempt++ {
		if w.jobs != nil {
			_ = w.jobs.MarkRunning(
				ctx,
				request.JobID,
				attempt,
			)
		}

		_, err = w.syncService.Sync(
			ctx,
			request.RepositoryID,
		)

		if err == nil {
			w.metrics.Succeeded.Add(1)

			if w.jobs != nil {
				_ = w.jobs.MarkSucceeded(
					ctx,
					request.JobID,
					attempt,
				)
			}

			return nil
		}

		if ctx.Err() != nil {
			w.metrics.Failed.Add(1)
			return ctx.Err()
		}

		if attempt == w.policy.MaxAttempts {
			w.metrics.Failed.Add(1)

			if w.jobs != nil {
				_ = w.jobs.MarkFailed(
					ctx,
					request.JobID,
					attempt,
					"repository_sync_failed",
					err.Error(),
				)
			}

			break
		}

		w.metrics.Retried.Add(1)

		timer := time.NewTimer(w.policy.Delay)

		select {
		case <-timer.C:

		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}

	return err
}

func (w *RepositorySyncWorker) Metrics() RepositorySyncMetricsSnapshot {
	return w.metrics.Snapshot()
}

var _ RepositorySyncQueue = (*RepositorySyncWorker)(nil)
