package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type retryImportService struct {
	attempts int
	failures int
	err      error
}

func (s *retryImportService) ImportRepositories(
	ctx context.Context,
	requesterID string,
	projectID string,
) ([]*domain.Repository, error) {
	s.attempts++
	if s.attempts <= s.failures {
		return nil, s.err
	}
	return []*domain.Repository{}, nil
}

func TestGitHubRepositoryImportWorkerRetriesUntilSuccess(t *testing.T) {
	importer := &retryImportService{
		failures: 2,
		err:      errors.New("temporary github failure"),
	}
	worker := NewGitHubRepositoryImportWorker(
		importer,
		1,
		GitHubRepositoryImportRetryPolicy{
			MaxAttempts: 3,
		},
	)

	err := worker.runWithRetry(
		context.Background(),
		GitHubRepositoryImportRequest{
			RequesterID: "user-123",
			ProjectID:   "project-123",
		},
	)

	if err != nil {
		t.Fatalf("expected retry to succeed: %v", err)
	}
	if importer.attempts != 3 {
		t.Fatalf("expected three attempts, got %d", importer.attempts)
	}
	metrics := worker.Metrics()
	if metrics.Retried != 2 || metrics.Succeeded != 1 {
		t.Fatalf("unexpected retry metrics: %+v", metrics)
	}
}

func TestGitHubRepositoryImportWorkerStopsAfterMaxAttempts(t *testing.T) {
	importer := &retryImportService{
		failures: 5,
		err:      errors.New("persistent github failure"),
	}
	worker := NewGitHubRepositoryImportWorker(
		importer,
		1,
		GitHubRepositoryImportRetryPolicy{
			MaxAttempts: 3,
		},
	)

	err := worker.runWithRetry(context.Background(), GitHubRepositoryImportRequest{})

	if !errors.Is(err, importer.err) {
		t.Fatalf("expected persistent error, got %v", err)
	}
	if importer.attempts != 3 {
		t.Fatalf("expected three attempts, got %d", importer.attempts)
	}
	metrics := worker.Metrics()
	if metrics.Retried != 2 || metrics.Failed != 1 {
		t.Fatalf("unexpected failure metrics: %+v", metrics)
	}
}

func TestGitHubRepositoryImportWorkerStopsOnCancellation(t *testing.T) {
	importer := &retryImportService{
		failures: 5,
		err:      errors.New("temporary github failure"),
	}
	worker := NewGitHubRepositoryImportWorker(
		importer,
		1,
		GitHubRepositoryImportRetryPolicy{
			MaxAttempts: 3,
		},
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := worker.runWithRetry(ctx, GitHubRepositoryImportRequest{})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if importer.attempts != 1 {
		t.Fatalf("expected one attempt, got %d", importer.attempts)
	}
}

func TestGitHubRepositoryImportWorkerEnqueue(t *testing.T) {
	worker := NewGitHubRepositoryImportWorker(
		&retryImportService{},
		1,
		GitHubRepositoryImportRetryPolicy{},
	)

	request := GitHubRepositoryImportRequest{
		RequesterID: "user-123",
		ProjectID:   "project-123",
	}
	if _, err := worker.Enqueue(context.Background(), request); err != nil {
		t.Fatalf("enqueue import: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := worker.Enqueue(ctx, GitHubRepositoryImportRequest{}); !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf("expected canceled enqueue, got %v", err)
	}
}

type shutdownJobRepository struct {
	interruptedBefore time.Time
	interruptedCode   string
	interruptedCount  int64
	interruptedErr    error
	markedFailed      bool
	jobID             string
	attempts          int
	code              string
	message           string
}

func (r *shutdownJobRepository) Create(
	ctx context.Context,
	job *domain.GitHubRepositoryImportJob,
) error {
	return nil
}

func (r *shutdownJobRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.GitHubRepositoryImportJob, error) {
	return nil, nil
}

func (r *shutdownJobRepository) MarkRunning(
	ctx context.Context,
	id string,
	attempts int,
) error {
	return nil
}

func (r *shutdownJobRepository) MarkSucceeded(
	ctx context.Context,
	id string,
	attempts int,
) error {
	return nil
}

func (r *shutdownJobRepository) MarkFailed(
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

func TestGitHubRepositoryImportWorkerMarksQueuedJobFailedOnShutdown(t *testing.T) {
	importer := &retryImportService{}
	jobs := &shutdownJobRepository{}

	worker := NewGitHubRepositoryImportWorkerWithStore(
		importer,
		1,
		GitHubRepositoryImportRetryPolicy{
			MaxAttempts: 3,
		},
		jobs,
	)

	job, err := worker.Enqueue(
		context.Background(),
		GitHubRepositoryImportRequest{
			RequesterID: "user-123",
			ProjectID:   "project-123",
		},
	)
	if err != nil {
		t.Fatalf("enqueue import: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker.Start(ctx)

	if !jobs.markedFailed {
		t.Fatal("expected queued job to be marked failed")
	}

	if jobs.jobID != job.ID {
		t.Fatalf("expected job ID %q, got %q", job.ID, jobs.jobID)
	}

	if jobs.attempts != 0 {
		t.Fatalf("expected zero attempts, got %d", jobs.attempts)
	}

	if jobs.code != domain.GitHubRepositoryImportJobFailureCodeWorkerShutdown {
		t.Fatalf("unexpected failure code: %q", jobs.code)
	}

	if jobs.message != "worker shutting down before job was processed" {
		t.Fatalf("unexpected failure message: %q", jobs.message)
	}

	if importer.attempts != 0 {
		t.Fatalf("expected importer not to run, got %d attempts", importer.attempts)
	}
}
func (r *shutdownJobRepository) FailInterrupted(
	ctx context.Context,
	updatedBefore time.Time,
	code string,
	message string,
) (int64, error) {
	r.interruptedBefore = updatedBefore
	r.interruptedCode = code
	return r.interruptedCount, r.interruptedErr
}

func TestGitHubRepositoryImportWorkerRecoversInterruptedJobs(t *testing.T) {
	jobs := &shutdownJobRepository{interruptedCount: 2}
	worker := NewGitHubRepositoryImportWorkerWithStore(
		&retryImportService{}, 1,
		GitHubRepositoryImportRetryPolicy{MaxAttempts: 1}, jobs,
	)
	startedAt := time.Now()

	recovered, err := worker.RecoverInterruptedJobs(context.Background(), startedAt)
	if err != nil {
		t.Fatalf("recover interrupted jobs: %v", err)
	}
	if recovered != 2 {
		t.Fatalf("expected 2 recovered jobs, got %d", recovered)
	}
	if !jobs.interruptedBefore.Equal(startedAt) {
		t.Fatal("expected recovery to use the process start time as cutoff")
	}
	if jobs.interruptedCode != domain.GitHubRepositoryImportJobFailureCodeInterrupted {
		t.Fatalf("unexpected failure code: %q", jobs.interruptedCode)
	}
	if worker.Metrics().Failed != 2 {
		t.Fatalf("expected failed metric 2, got %d", worker.Metrics().Failed)
	}
}

func TestGitHubRepositoryImportWorkerRecoverPropagatesError(t *testing.T) {
	jobs := &shutdownJobRepository{interruptedErr: errors.New("db down")}
	worker := NewGitHubRepositoryImportWorkerWithStore(
		&retryImportService{}, 1,
		GitHubRepositoryImportRetryPolicy{MaxAttempts: 1}, jobs,
	)
	if _, err := worker.RecoverInterruptedJobs(context.Background(), time.Now()); err == nil {
		t.Fatal("expected error to propagate")
	}
}

func TestGitHubRepositoryImportWorkerRecoverWithoutStore(t *testing.T) {
	worker := NewGitHubRepositoryImportWorker(
		&retryImportService{}, 1,
		GitHubRepositoryImportRetryPolicy{MaxAttempts: 1},
	)
	recovered, err := worker.RecoverInterruptedJobs(context.Background(), time.Now())
	if err != nil || recovered != 0 {
		t.Fatalf("expected no-op, got %d, %v", recovered, err)
	}
}
