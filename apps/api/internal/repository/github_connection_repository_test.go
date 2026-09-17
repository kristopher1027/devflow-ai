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

func setupGitHubConnectionRepository(
	t *testing.T,
) (context.Context, *database.DB, GitHubConnectionRepository, string) {
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
	if _, err := db.Pool.Exec(
		ctx,
		`INSERT INTO users (id, email) VALUES ($1, $2)`,
		userID,
		"github-connection-"+uuid.NewString()+"@example.com",
	); err != nil {
		db.Close()
		t.Fatalf("insert test user: %v", err)
	}

	workspaceID := uuid.NewString()
	if _, err := db.Pool.Exec(
		ctx,
		`INSERT INTO workspaces (id, owner_id, name) VALUES ($1, $2, $3)`,
		workspaceID,
		userID,
		"GitHub Connection Test Workspace",
	); err != nil {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
		db.Close()
		t.Fatalf("insert test workspace: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM workspaces WHERE id = $1",
			workspaceID,
		)
		if err != nil {
			t.Errorf("cleanup test workspace: %v", err)
		}

		_, err = db.Pool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("cleanup test user: %v", err)
		}
		db.Close()
	})

	return ctx, db, NewGitHubConnectionRepository(db), workspaceID
}

func testGitHubConnection(workspaceID string) *domain.GitHubConnection {
	now := time.Now()

	return &domain.GitHubConnection{
		ID:             uuid.NewString(),
		WorkspaceID:    workspaceID,
		InstallationID: "installation-" + uuid.NewString(),
		AccountLogin:   "devflow-org",
		Status:         domain.GitHubConnectionStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func TestGitHubConnectionRepositoryCreateAndFind(t *testing.T) {
	ctx, _, repo, workspaceID := setupGitHubConnectionRepository(t)
	connection := testGitHubConnection(workspaceID)

	if err := repo.Create(ctx, connection); err != nil {
		t.Fatalf("create github connection: %v", err)
	}

	byWorkspace, err := repo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		t.Fatalf("find github connection by workspace: %v", err)
	}

	if byWorkspace.ID != connection.ID {
		t.Fatalf("expected connection ID %s, got %s", connection.ID, byWorkspace.ID)
	}

	byInstallation, err := repo.FindByInstallationID(
		ctx,
		connection.InstallationID,
	)
	if err != nil {
		t.Fatalf("find github connection by installation: %v", err)
	}

	if byInstallation.WorkspaceID != workspaceID {
		t.Fatalf(
			"expected workspace ID %s, got %s",
			workspaceID,
			byInstallation.WorkspaceID,
		)
	}

	if byInstallation.Status != domain.GitHubConnectionStatusPending {
		t.Fatalf(
			"expected status %s, got %s",
			domain.GitHubConnectionStatusPending,
			byInstallation.Status,
		)
	}
}

func TestGitHubConnectionRepositoryFindNotFound(t *testing.T) {
	ctx, _, repo, _ := setupGitHubConnectionRepository(t)

	_, err := repo.FindByWorkspaceID(ctx, uuid.NewString())
	if !errors.Is(err, ErrGitHubConnectionNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}

	_, err = repo.FindByInstallationID(ctx, "missing-installation")
	if !errors.Is(err, ErrGitHubConnectionNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGitHubConnectionRepositoryUpdateStatus(t *testing.T) {
	ctx, _, repo, workspaceID := setupGitHubConnectionRepository(t)
	connection := testGitHubConnection(workspaceID)

	if err := repo.Create(ctx, connection); err != nil {
		t.Fatalf("create github connection: %v", err)
	}

	if err := repo.UpdateStatus(
		ctx,
		connection.ID,
		domain.GitHubConnectionStatusActive,
	); err != nil {
		t.Fatalf("update github connection status: %v", err)
	}

	updated, err := repo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		t.Fatalf("find updated github connection: %v", err)
	}

	if updated.Status != domain.GitHubConnectionStatusActive {
		t.Fatalf(
			"expected status %s, got %s",
			domain.GitHubConnectionStatusActive,
			updated.Status,
		)
	}
}

func TestGitHubConnectionRepositoryUpdateStatusValidation(t *testing.T) {
	ctx, _, repo, _ := setupGitHubConnectionRepository(t)

	if err := repo.UpdateStatus(ctx, uuid.NewString(), "unknown"); !errors.Is(
		err,
		domain.ErrGitHubConnectionStatusInvalid,
	) {
		t.Fatalf("expected invalid status error, got %v", err)
	}

	if err := repo.UpdateStatus(
		ctx,
		uuid.NewString(),
		domain.GitHubConnectionStatusActive,
	); !errors.Is(err, ErrGitHubConnectionNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
