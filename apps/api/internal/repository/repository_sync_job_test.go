package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

func setupRepositorySyncJobRepository(
	t *testing.T,
) (context.Context, *database.DB, RepositorySyncJobRepository, string) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()
	repositoryID := uuid.NewString()

	if _, err := db.Pool.Exec(
		ctx,
		`INSERT INTO users (id, email) VALUES ($1, $2)`,
		userID,
		userID+"@example.com",
	); err != nil {
		db.Close()
		t.Fatalf("insert user: %v", err)
	}

	if _, err := db.Pool.Exec(
		ctx,
		`INSERT INTO workspaces (id, owner_id, name) VALUES ($1, $2, $3)`,
		workspaceID,
		userID,
		"Sync Job Workspace",
	); err != nil {
		_, _ = db.Pool.Exec(
			ctx,
			`DELETE FROM users WHERE id = $1`,
			userID,
		)
		db.Close()
		t.Fatalf("insert workspace: %v", err)
	}

	if _, err := db.Pool.Exec(
		ctx,
		`INSERT INTO projects (id, workspace_id, name, created_by)
		 VALUES ($1, $2, $3, $4)`,
		projectID,
		workspaceID,
		"Sync Job Project",
		userID,
	); err != nil {
		_, _ = db.Pool.Exec(
			ctx,
			`DELETE FROM workspaces WHERE id = $1`,
			workspaceID,
		)
		_, _ = db.Pool.Exec(
			ctx,
			`DELETE FROM users WHERE id = $1`,
			userID,
		)
		db.Close()
		t.Fatalf("insert project: %v", err)
	}

	if _, err := db.Pool.Exec(
		ctx,
		`INSERT INTO repositories (
			id,
			project_id,
			provider,
			external_id,
			owner,
			name,
			full_name,
			default_branch,
			html_url,
			clone_url
		)
		VALUES (
			$1,
			$2,
			'github',
			$3,
			$4,
			$5,
			$6,
			'main',
			$7,
			$8
		)`,
		repositoryID,
		projectID,
		repositoryID,
		"test-owner",
		"test-repository",
		"test-owner/test-repository",
		"https://github.com/test-owner/test-repository",
		"https://github.com/test-owner/test-repository.git",
	); err != nil {
		_, _ = db.Pool.Exec(
			ctx,
			`DELETE FROM workspaces WHERE id = $1`,
			workspaceID,
		)
		_, _ = db.Pool.Exec(
			ctx,
			`DELETE FROM users WHERE id = $1`,
			userID,
		)
		db.Close()
		t.Fatalf("insert repository: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(
			context.Background(),
			`DELETE FROM workspaces WHERE id = $1`,
			workspaceID,
		)
		_, _ = db.Pool.Exec(
			context.Background(),
			`DELETE FROM users WHERE id = $1`,
			userID,
		)
		db.Close()
	})

	return ctx, db, NewRepositorySyncJobRepository(db), repositoryID
}

func TestRepositorySyncJobRepositoryLifecycle(t *testing.T) {
	ctx, _, repo, repositoryID := setupRepositorySyncJobRepository(t)

	now := time.Now()

	job := &domain.RepositorySyncJob{
		ID:           uuid.NewString(),
		RepositoryID: repositoryID,
		Status:       domain.RepositorySyncJobStatusPending,
		Attempts:     0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("create sync job: %v", err)
	}

	if err := repo.MarkRunning(ctx, job.ID, 1); err != nil {
		t.Fatalf("mark running: %v", err)
	}

	if err := repo.MarkSucceeded(ctx, job.ID, 1); err != nil {
		t.Fatalf("mark succeeded: %v", err)
	}

	stored, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("find sync job: %v", err)
	}

	if stored.Status != domain.RepositorySyncJobStatusSucceeded {
		t.Fatalf("expected succeeded status, got %q", stored.Status)
	}

	if stored.Attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", stored.Attempts)
	}

	if stored.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}

	if stored.FailureCode != nil {
		t.Fatal("expected failure code to be nil")
	}

	if stored.FailureMessage != nil {
		t.Fatal("expected failure message to be nil")
	}
}

func TestRepositorySyncJobRepositoryFailure(t *testing.T) {
	ctx, _, repo, repositoryID := setupRepositorySyncJobRepository(t)

	now := time.Now()

	job := &domain.RepositorySyncJob{
		ID:           uuid.NewString(),
		RepositoryID: repositoryID,
		Status:       domain.RepositorySyncJobStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("create sync job: %v", err)
	}

	if err := repo.MarkRunning(ctx, job.ID, 1); err != nil {
		t.Fatalf("mark running: %v", err)
	}

	if err := repo.MarkFailed(
		ctx,
		job.ID,
		1,
		"sync_failed",
		"repository synchronization failed",
	); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	stored, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("find sync job: %v", err)
	}

	if stored.Status != domain.RepositorySyncJobStatusFailed {
		t.Fatalf("expected failed status, got %q", stored.Status)
	}

	if stored.FailureCode == nil ||
		*stored.FailureCode != "sync_failed" {
		t.Fatal("expected failure code")
	}

	if stored.FailureMessage == nil ||
		*stored.FailureMessage != "repository synchronization failed" {
		t.Fatal("expected failure message")
	}

	if stored.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}
}

func TestRepositorySyncJobRepositoryListByRepositoryID(t *testing.T) {
	ctx, _, repo, repositoryID := setupRepositorySyncJobRepository(t)

	now := time.Now()

	for i := 0; i < 2; i++ {
		job := &domain.RepositorySyncJob{
			ID:           uuid.NewString(),
			RepositoryID: repositoryID,
			Status:       domain.RepositorySyncJobStatusPending,
			CreatedAt:    now.Add(time.Duration(i) * time.Second),
			UpdatedAt:    now.Add(time.Duration(i) * time.Second),
		}

		if err := repo.Create(ctx, job); err != nil {
			t.Fatalf("create sync job %d: %v", i, err)
		}
	}

	jobs, err := repo.ListByRepositoryID(ctx, repositoryID)
	if err != nil {
		t.Fatalf("list sync jobs: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 sync jobs, got %d", len(jobs))
	}

	if jobs[0].CreatedAt.Before(jobs[1].CreatedAt) {
		t.Fatal("expected newest sync job first")
	}
}

func TestRepositorySyncJobRepositoryNotFound(t *testing.T) {
	ctx, _, repo, _ := setupRepositorySyncJobRepository(t)

	_, err := repo.FindByID(ctx, uuid.NewString())

	if !errors.Is(err, ErrRepositorySyncJobNotFound) {
		t.Fatalf(
			"expected not found error, got %v",
			err,
		)
	}
}
