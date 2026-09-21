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

func setupRepositorySnapshotRepository(
	t *testing.T,
) (context.Context, *database.DB, RepositorySnapshotRepository, string) {
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
		`INSERT INTO workspaces (id, owner_id, name)
		 VALUES ($1, $2, $3)`,
		workspaceID,
		userID,
		"Snapshot Test Workspace",
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
		`INSERT INTO projects (
			id,
			workspace_id,
			name,
			created_by
		)
		VALUES ($1, $2, $3, $4)`,
		projectID,
		workspaceID,
		"Snapshot Test Project",
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
		"snapshot-test-repository",
		"test-owner/snapshot-test-repository",
		"https://github.com/test-owner/snapshot-test-repository",
		"https://github.com/test-owner/snapshot-test-repository.git",
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

	return ctx, db, NewRepositorySnapshotRepository(db), repositoryID
}

func TestRepositorySnapshotRepositoryCreateAndFind(t *testing.T) {
	ctx, _, repo, repositoryID := setupRepositorySnapshotRepository(t)

	now := time.Now()

	snapshot := &domain.RepositorySnapshot{
		ID:           uuid.NewString(),
		RepositoryID: repositoryID,
		CommitSHA:    "abc123def456",
		Branch:       "main",
		CreatedAt:    now,
	}

	if err := repo.Create(ctx, snapshot); err != nil {
		t.Fatalf("create repository snapshot: %v", err)
	}

	stored, err := repo.FindByID(ctx, snapshot.ID)
	if err != nil {
		t.Fatalf("find repository snapshot: %v", err)
	}

	if stored.ID != snapshot.ID {
		t.Fatalf("expected ID %q, got %q", snapshot.ID, stored.ID)
	}

	if stored.RepositoryID != repositoryID {
		t.Fatalf(
			"expected repository ID %q, got %q",
			repositoryID,
			stored.RepositoryID,
		)
	}

	if stored.CommitSHA != snapshot.CommitSHA {
		t.Fatalf(
			"expected commit SHA %q, got %q",
			snapshot.CommitSHA,
			stored.CommitSHA,
		)
	}

	if stored.Branch != snapshot.Branch {
		t.Fatalf(
			"expected branch %q, got %q",
			snapshot.Branch,
			stored.Branch,
		)
	}
}

func TestRepositorySnapshotRepositoryFindNotFound(t *testing.T) {
	ctx, _, repo, _ := setupRepositorySnapshotRepository(t)

	_, err := repo.FindByID(ctx, uuid.NewString())
	if !errors.Is(err, ErrRepositorySnapshotNotFound) {
		t.Fatalf(
			"expected ErrRepositorySnapshotNotFound, got %v",
			err,
		)
	}
}

func TestRepositorySnapshotRepositoryListByRepositoryID(
	t *testing.T,
) {
	ctx, _, repo, repositoryID := setupRepositorySnapshotRepository(t)

	now := time.Now()

	first := &domain.RepositorySnapshot{
		ID:           uuid.NewString(),
		RepositoryID: repositoryID,
		CommitSHA:    "commit-one",
		Branch:       "main",
		CreatedAt:    now.Add(-2 * time.Hour),
	}

	second := &domain.RepositorySnapshot{
		ID:           uuid.NewString(),
		RepositoryID: repositoryID,
		CommitSHA:    "commit-two",
		Branch:       "main",
		CreatedAt:    now.Add(-1 * time.Hour),
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first snapshot: %v", err)
	}

	if err := repo.Create(ctx, second); err != nil {
		t.Fatalf("create second snapshot: %v", err)
	}

	snapshots, err := repo.ListByRepositoryID(ctx, repositoryID)
	if err != nil {
		t.Fatalf("list repository snapshots: %v", err)
	}

	if len(snapshots) != 2 {
		t.Fatalf(
			"expected 2 snapshots, got %d",
			len(snapshots),
		)
	}

	if snapshots[0].ID != second.ID {
		t.Fatalf(
			"expected newest snapshot first, got %q",
			snapshots[0].ID,
		)
	}

	if snapshots[1].ID != first.ID {
		t.Fatalf(
			"expected oldest snapshot second, got %q",
			snapshots[1].ID,
		)
	}
}

func TestRepositorySnapshotRepositoryDuplicateCommit(
	t *testing.T,
) {
	ctx, _, repo, repositoryID := setupRepositorySnapshotRepository(t)

	now := time.Now()

	first := &domain.RepositorySnapshot{
		ID:           uuid.NewString(),
		RepositoryID: repositoryID,
		CommitSHA:    "duplicate-commit",
		Branch:       "main",
		CreatedAt:    now,
	}

	second := &domain.RepositorySnapshot{
		ID:           uuid.NewString(),
		RepositoryID: repositoryID,
		CommitSHA:    "duplicate-commit",
		Branch:       "main",
		CreatedAt:    now.Add(time.Second),
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first snapshot: %v", err)
	}

	if err := repo.Create(ctx, second); err == nil {
		t.Fatal("expected duplicate commit to be rejected")
	}
}
