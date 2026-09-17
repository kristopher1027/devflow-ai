package service

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type GitHubRepositoryImportRequest struct {
	JobID       string
	RequesterID string
	ProjectID   string
}

type GitHubRepositoryImportQueue interface {
	Enqueue(
		ctx context.Context,
		request GitHubRepositoryImportRequest,
	) (*domain.GitHubRepositoryImportJob, error)
}

type GitHubRepositoryImportMetrics struct {
	Queued    atomic.Uint64
	Retried   atomic.Uint64
	Succeeded atomic.Uint64
	Failed    atomic.Uint64
}

type GitHubRepositoryImportMetricsSnapshot struct {
	Queued    uint64
	Retried   uint64
	Succeeded uint64
	Failed    uint64
}

func (m *GitHubRepositoryImportMetrics) Snapshot() GitHubRepositoryImportMetricsSnapshot {
	return GitHubRepositoryImportMetricsSnapshot{
		Queued: m.Queued.Load(), Retried: m.Retried.Load(),
		Succeeded: m.Succeeded.Load(), Failed: m.Failed.Load(),
	}
}

type GitHubRepositoryImportRetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
}

type GitHubRepositoryImportWorker struct {
	importer GitHubRepositoryImportService
	requests chan GitHubRepositoryImportRequest
	policy   GitHubRepositoryImportRetryPolicy
	jobs     repository.GitHubRepositoryImportJobRepository
	metrics  *GitHubRepositoryImportMetrics
}

func NewGitHubRepositoryImportWorker(
	importer GitHubRepositoryImportService,
	queueSize int,
	policy GitHubRepositoryImportRetryPolicy,
) *GitHubRepositoryImportWorker {
	return NewGitHubRepositoryImportWorkerWithStore(importer, queueSize, policy, nil)
}

func NewGitHubRepositoryImportWorkerWithStore(
	importer GitHubRepositoryImportService,
	queueSize int,
	policy GitHubRepositoryImportRetryPolicy,
	jobs repository.GitHubRepositoryImportJobRepository,
) *GitHubRepositoryImportWorker {
	if queueSize < 1 {
		queueSize = 1
	}
	if policy.MaxAttempts < 1 {
		policy.MaxAttempts = 1
	}

	return &GitHubRepositoryImportWorker{
		importer: importer,
		requests: make(chan GitHubRepositoryImportRequest, queueSize),
		policy:   policy,
		jobs:     jobs,
		metrics:  &GitHubRepositoryImportMetrics{},
	}
}

func (w *GitHubRepositoryImportWorker) Enqueue(
	ctx context.Context,
	request GitHubRepositoryImportRequest,
) (*domain.GitHubRepositoryImportJob, error) {
	job := &domain.GitHubRepositoryImportJob{
		ID: uuid.NewString(), RequesterID: request.RequesterID,
		ProjectID: request.ProjectID, Status: domain.GitHubRepositoryImportJobStatusPending,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
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

func (w *GitHubRepositoryImportWorker) Start(ctx context.Context) {
	for {
		select {
		case request := <-w.requests:
			_ = w.runWithRetry(ctx, request)
		case <-ctx.Done():
			return
		}
	}
}

func (w *GitHubRepositoryImportWorker) runWithRetry(
	ctx context.Context,
	request GitHubRepositoryImportRequest,
) error {
	var err error
	for attempt := 1; attempt <= w.policy.MaxAttempts; attempt++ {
		if w.jobs != nil {
			_ = w.jobs.MarkRunning(ctx, request.JobID, attempt)
		}
		_, err = w.importer.ImportRepositories(
			ctx,
			request.RequesterID,
			request.ProjectID,
		)
		if err == nil {
			w.metrics.Succeeded.Add(1)
			if w.jobs != nil {
				_ = w.jobs.MarkSucceeded(ctx, request.JobID, attempt)
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
				_ = w.jobs.MarkFailed(ctx, request.JobID, attempt, "import_failed", err.Error())
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

func (w *GitHubRepositoryImportWorker) Metrics() GitHubRepositoryImportMetricsSnapshot {
	return w.metrics.Snapshot()
}

var _ GitHubRepositoryImportQueue = (*GitHubRepositoryImportWorker)(nil)
