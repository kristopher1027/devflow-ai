package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type retryRepositorySyncService struct {
	attempts int
	failures int
	err      error
}

func (s *retryRepositorySyncService) Sync(
	ctx context.Context,
	repositoryID string,
) (*domain.RepositorySnapshot, error) {
	s.attempts++

	if s.attempts <= s.failures {
		return nil, s.err
	}

	return &domain.RepositorySnapshot{
		ID:           "snapshot-123",
		RepositoryID: repositoryID,
		CommitSHA:    "abc123",
		Branch:       "main",
	}, nil
}

func TestRepositorySyncWorkerRetriesUntilSuccess(t *testing.T) {
	syncService := &retryRepositorySyncService{
		failures: 2,
		err:      errors.New("temporary sync failure"),
	}

	worker := NewRepositorySyncWorker(
		syncService,
		1,
		RepositorySyncRetryPolicy{
			MaxAttempts: 3,
		},
	)

	err := worker.runWithRetry(
		context.Background(),
		RepositorySyncRequest{
			RepositoryID: "repository-123",
		},
	)

	if err != nil {
		t.Fatalf("expected retry to succeed: %v", err)
	}

	if syncService.attempts != 3 {
		t.Fatalf(
			"expected three attempts, got %d",
			syncService.attempts,
		)
	}

	metrics := worker.Metrics()

	if metrics.Retried != 2 || metrics.Succeeded != 1 {
		t.Fatalf(
			"unexpected retry metrics: %+v",
			metrics,
		)
	}
}

func TestRepositorySyncWorkerStopsAfterMaxAttempts(t *testing.T) {
	syncService := &retryRepositorySyncService{
		failures: 5,
		err:      errors.New("persistent sync failure"),
	}

	worker := NewRepositorySyncWorker(
		syncService,
		1,
		RepositorySyncRetryPolicy{
			MaxAttempts: 3,
		},
	)

	err := worker.runWithRetry(
		context.Background(),
		RepositorySyncRequest{},
	)

	if !errors.Is(err, syncService.err) {
		t.Fatalf(
			"expected persistent error, got %v",
			err,
		)
	}

	if syncService.attempts != 3 {
		t.Fatalf(
			"expected three attempts, got %d",
			syncService.attempts,
		)
	}

	metrics := worker.Metrics()

	if metrics.Retried != 2 || metrics.Failed != 1 {
		t.Fatalf(
			"unexpected failure metrics: %+v",
			metrics,
		)
	}
}

func TestRepositorySyncWorkerStopsOnCancellation(t *testing.T) {
	syncService := &retryRepositorySyncService{
		failures: 5,
		err:      errors.New("temporary sync failure"),
	}

	worker := NewRepositorySyncWorker(
		syncService,
		1,
		RepositorySyncRetryPolicy{
			MaxAttempts: 3,
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := worker.runWithRetry(
		ctx,
		RepositorySyncRequest{},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected cancellation, got %v",
			err,
		)
	}

	if syncService.attempts != 1 {
		t.Fatalf(
			"expected one attempt, got %d",
			syncService.attempts,
		)
	}
}

func TestRepositorySyncWorkerEnqueue(t *testing.T) {
	worker := NewRepositorySyncWorker(
		&retryRepositorySyncService{},
		1,
		RepositorySyncRetryPolicy{},
	)

	request := RepositorySyncRequest{
		RepositoryID: "repository-123",
	}

	job, err := worker.Enqueue(
		context.Background(),
		request,
	)

	if err != nil {
		t.Fatalf("enqueue repository sync: %v", err)
	}

	if job.ID == "" {
		t.Fatal("expected job ID")
	}

	if job.RepositoryID != request.RepositoryID {
		t.Fatalf(
			"expected repository ID %q, got %q",
			request.RepositoryID,
			job.RepositoryID,
		)
	}

	if job.Status != domain.RepositorySyncJobStatusPending {
		t.Fatalf(
			"expected pending status, got %q",
			job.Status,
		)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := worker.Enqueue(
		ctx,
		RepositorySyncRequest{},
	); !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected canceled enqueue, got %v",
			err,
		)
	}
}

type repositorySyncJobStore struct {
	markedRunning   bool
	markedSucceeded bool
	markedFailed    bool

	jobID    string
	attempts int
	code     string
	message  string
}

func (r *repositorySyncJobStore) Create(
	ctx context.Context,
	job *domain.RepositorySyncJob,
) error {
	return nil
}

func (r *repositorySyncJobStore) FindByID(
	ctx context.Context,
	id string,
) (*domain.RepositorySyncJob, error) {
	return nil, nil
}

func (r *repositorySyncJobStore) ListByRepositoryID(
	ctx context.Context,
	repositoryID string,
) ([]*domain.RepositorySyncJob, error) {
	return nil, nil
}

func (r *repositorySyncJobStore) MarkRunning(
	ctx context.Context,
	id string,
	attempts int,
) error {
	r.markedRunning = true
	r.jobID = id
	r.attempts = attempts
	return nil
}

func (r *repositorySyncJobStore) MarkSucceeded(
	ctx context.Context,
	id string,
	attempts int,
) error {
	r.markedSucceeded = true
	r.jobID = id
	r.attempts = attempts
	return nil
}

func (r *repositorySyncJobStore) MarkFailed(
	ctx context.Context,
	id string,
	attempts int,
	code string,
	message string,
) error {
	r.markedFailed = true
	r.jobID = id
	r.attempts = attempts
	r.code = code
	r.message = message
	return nil
}

func TestRepositorySyncWorkerMarksJobSucceeded(t *testing.T) {
	syncService := &retryRepositorySyncService{}

	jobs := &repositorySyncJobStore{}

	worker := NewRepositorySyncWorkerWithStore(
		syncService,
		1,
		RepositorySyncRetryPolicy{
			MaxAttempts: 3,
		},
		jobs,
	)

	job, err := worker.Enqueue(
		context.Background(),
		RepositorySyncRequest{
			RepositoryID: "repository-123",
		},
	)

	if err != nil {
		t.Fatalf("enqueue repository sync: %v", err)
	}

	err = worker.runWithRetry(
		context.Background(),
		RepositorySyncRequest{
			JobID:        job.ID,
			RepositoryID: "repository-123",
		},
	)

	if err != nil {
		t.Fatalf("expected sync to succeed: %v", err)
	}

	if !jobs.markedRunning {
		t.Fatal("expected job to be marked running")
	}

	if !jobs.markedSucceeded {
		t.Fatal("expected job to be marked succeeded")
	}

	if jobs.jobID != job.ID {
		t.Fatalf(
			"expected job ID %q, got %q",
			job.ID,
			jobs.jobID,
		)
	}

	if jobs.attempts != 1 {
		t.Fatalf(
			"expected one attempt, got %d",
			jobs.attempts,
		)
	}
}

func TestRepositorySyncWorkerMarksJobFailedAfterMaxAttempts(
	t *testing.T,
) {
	syncService := &retryRepositorySyncService{
		failures: 5,
		err:      errors.New("persistent sync failure"),
	}

	jobs := &repositorySyncJobStore{}

	worker := NewRepositorySyncWorkerWithStore(
		syncService,
		1,
		RepositorySyncRetryPolicy{
			MaxAttempts: 3,
		},
		jobs,
	)

	job, err := worker.Enqueue(
		context.Background(),
		RepositorySyncRequest{
			RepositoryID: "repository-123",
		},
	)

	if err != nil {
		t.Fatalf("enqueue repository sync: %v", err)
	}

	err = worker.runWithRetry(
		context.Background(),
		RepositorySyncRequest{
			JobID:        job.ID,
			RepositoryID: "repository-123",
		},
	)

	if !errors.Is(err, syncService.err) {
		t.Fatalf(
			"expected persistent error, got %v",
			err,
		)
	}

	if !jobs.markedFailed {
		t.Fatal("expected job to be marked failed")
	}

	if jobs.jobID != job.ID {
		t.Fatalf(
			"expected job ID %q, got %q",
			job.ID,
			jobs.jobID,
		)
	}

	if jobs.attempts != 3 {
		t.Fatalf(
			"expected three attempts, got %d",
			jobs.attempts,
		)
	}

	if jobs.code != "repository_sync_failed" {
		t.Fatalf(
			"unexpected failure code: %q",
			jobs.code,
		)
	}

	if jobs.message != syncService.err.Error() {
		t.Fatalf(
			"unexpected failure message: %q",
			jobs.message,
		)
	}
}

func TestRepositorySyncWorkerMarksQueuedJobFailedOnShutdown(
	t *testing.T,
) {
	syncService := &retryRepositorySyncService{}

	jobs := &repositorySyncJobStore{}

	worker := NewRepositorySyncWorkerWithStore(
		syncService,
		1,
		RepositorySyncRetryPolicy{
			MaxAttempts: 3,
		},
		jobs,
	)

	job, err := worker.Enqueue(
		context.Background(),
		RepositorySyncRequest{
			RepositoryID: "repository-123",
		},
	)

	if err != nil {
		t.Fatalf("enqueue repository sync: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker.Start(ctx)

	if !jobs.markedFailed {
		t.Fatal("expected queued job to be marked failed")
	}

	if jobs.jobID != job.ID {
		t.Fatalf(
			"expected job ID %q, got %q",
			job.ID,
			jobs.jobID,
		)
	}

	if jobs.attempts != 0 {
		t.Fatalf(
			"expected zero attempts, got %d",
			jobs.attempts,
		)
	}

	if jobs.code != domain.RepositorySyncJobFailureCodeWorkerShutdown {
		t.Fatalf(
			"unexpected failure code: %q",
			jobs.code,
		)
	}

	if jobs.message != "worker shutting down before job was processed" {
		t.Fatalf(
			"unexpected failure message: %q",
			jobs.message,
		)
	}

	if syncService.attempts != 0 {
		t.Fatalf(
			"expected sync service not to run, got %d attempts",
			syncService.attempts,
		)
	}
}
