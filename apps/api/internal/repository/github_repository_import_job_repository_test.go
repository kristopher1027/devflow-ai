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

func setupImportJobRepository(t *testing.T) (context.Context, *database.DB, GitHubRepositoryImportJobRepository, string, string) {
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
	if _, err := db.Pool.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2)`, userID, userID+"@example.com"); err != nil {
		db.Close()
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.Pool.Exec(ctx, `INSERT INTO workspaces (id, owner_id, name) VALUES ($1, $2, $3)`, workspaceID, userID, "Import Job Workspace"); err != nil {
		_, _ = db.Pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		db.Close()
		t.Fatalf("insert workspace: %v", err)
	}
	if _, err := db.Pool.Exec(ctx, `INSERT INTO projects (id, workspace_id, name, created_by) VALUES ($1, $2, $3, $4)`, projectID, workspaceID, "Import Job Project", userID); err != nil {
		_, _ = db.Pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, workspaceID)
		_, _ = db.Pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		db.Close()
		t.Fatalf("insert project: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), `DELETE FROM workspaces WHERE id = $1`, workspaceID)
		_, _ = db.Pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
		db.Close()
	})
	return ctx, db, NewGitHubRepositoryImportJobRepository(db), userID, projectID
}

func TestGitHubRepositoryImportJobRepositoryLifecycle(t *testing.T) {
	ctx, _, repo, userID, projectID := setupImportJobRepository(t)
	now := time.Now()
	job := &domain.GitHubRepositoryImportJob{
		ID: uuid.NewString(), RequesterID: userID, ProjectID: projectID,
		Status:    domain.GitHubRepositoryImportJobStatusPending,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("create import job: %v", err)
	}
	if err := repo.MarkRunning(ctx, job.ID, 1); err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if err := repo.MarkFailed(ctx, job.ID, 1, "github_unavailable", "GitHub API unavailable"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	stored, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("find import job: %v", err)
	}
	if stored.Status != domain.GitHubRepositoryImportJobStatusFailed || stored.Attempts != 1 {
		t.Fatalf("unexpected job state: %+v", stored)
	}
	if stored.FailureCode == nil || *stored.FailureCode != "github_unavailable" {
		t.Fatal("expected structured failure code")
	}
	if stored.FailureMessage == nil || *stored.FailureMessage != "GitHub API unavailable" {
		t.Fatal("expected structured failure message")
	}
}

func TestGitHubRepositoryImportJobRepositoryNotFound(t *testing.T) {
	ctx, _, repo, _, _ := setupImportJobRepository(t)
	_, err := repo.FindByID(ctx, uuid.NewString())
	if !errors.Is(err, ErrGitHubRepositoryImportJobNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
