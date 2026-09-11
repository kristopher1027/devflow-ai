package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

func TestWorkspaceRepositoryCreate(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := NewWorkspaceRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	email := "workspace-create-" + uuid.NewString() + "@example.com"

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO users (id, email)
		VALUES ($1, $2)
		`,
		userID,
		email,
	)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM workspaces WHERE owner_id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("cleanup test workspaces: %v", err)
		}

		_, err = db.Pool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("cleanup test user: %v", err)
		}
	})
	now := time.Now()

	workspace := &domain.Workspace{
		ID:        workspaceID,
		OwnerID:   userID,
		Name:      "Test Workspace",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = repo.Create(ctx, workspace)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	var (
		storedOwnerID string
		storedName    string
	)

	err = db.Pool.QueryRow(
		ctx,
		`
		SELECT owner_id, name
		FROM workspaces
		WHERE id = $1
		`,
		workspaceID,
	).Scan(
		&storedOwnerID,
		&storedName,
	)
	if err != nil {
		t.Fatalf("query created workspace: %v", err)
	}

	if storedOwnerID != userID {
		t.Fatalf(
			"expected owner ID %s, got %s",
			userID,
			storedOwnerID,
		)
	}

	if storedName != workspace.Name {
		t.Fatalf(
			"expected workspace name %s, got %s",
			workspace.Name,
			storedName,
		)
	}
}

func TestWorkspaceRepositoryFindByID(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := NewWorkspaceRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	email := "workspace-find-" + uuid.NewString() + "@example.com"

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO users (id, email)
		VALUES ($1, $2)
		`,
		userID,
		email,
	)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM workspaces WHERE owner_id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("cleanup test workspaces: %v", err)
		}

		_, err = db.Pool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("cleanup test user: %v", err)
		}
	})

	now := time.Now()

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO workspaces (
			id,
			owner_id,
			name,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
		`,
		workspaceID,
		userID,
		"Find Workspace",
		now,
		now,
	)
	if err != nil {
		t.Fatalf("insert test workspace: %v", err)
	}

	workspace, err := repo.FindByID(ctx, workspaceID)
	if err != nil {
		t.Fatalf("find workspace: %v", err)
	}

	if workspace.ID != workspaceID {
		t.Fatalf(
			"expected workspace ID %s, got %s",
			workspaceID,
			workspace.ID,
		)
	}

	if workspace.OwnerID != userID {
		t.Fatalf(
			"expected owner ID %s, got %s",
			userID,
			workspace.OwnerID,
		)
	}

	if workspace.Name != "Find Workspace" {
		t.Fatalf(
			"expected workspace name %q, got %q",
			"Find Workspace",
			workspace.Name,
		)
	}
}

func TestWorkspaceRepositoryFindByIDNotFound(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := NewWorkspaceRepository(db)

	_, err = repo.FindByID(
		ctx,
		uuid.NewString(),
	)

	if err != ErrWorkspaceNotFound {
		t.Fatalf(
			"expected ErrWorkspaceNotFound, got %v",
			err,
		)
	}
}

func TestWorkspaceRepositoryListByOwnerID(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := NewWorkspaceRepository(db)

	userID := uuid.NewString()
	email := "workspace-list-" + uuid.NewString() + "@example.com"

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO users (id, email)
		VALUES ($1, $2)
		`,
		userID,
		email,
	)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM workspaces WHERE owner_id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("cleanup test workspaces: %v", err)
		}

		_, err = db.Pool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("cleanup test user: %v", err)
		}
	})

	firstID := uuid.NewString()
	secondID := uuid.NewString()

	now := time.Now()

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO workspaces (
			id,
			owner_id,
			name,
			created_at,
			updated_at
		)
		VALUES
			($1, $2, $3, $4, $5),
			($6, $2, $7, $8, $9)
		`,
		firstID,
		userID,
		"First Workspace",
		now,
		now,
		secondID,
		"Second Workspace",
		now.Add(time.Second),
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("insert test workspaces: %v", err)
	}

	workspaces, err := repo.ListByOwnerID(ctx, userID)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}

	if len(workspaces) != 2 {
		t.Fatalf(
			"expected 2 workspaces, got %d",
			len(workspaces),
		)
	}

	if workspaces[0].ID != firstID {
		t.Fatalf(
			"expected first workspace ID %s, got %s",
			firstID,
			workspaces[0].ID,
		)
	}

	if workspaces[1].ID != secondID {
		t.Fatalf(
			"expected second workspace ID %s, got %s",
			secondID,
			workspaces[1].ID,
		)
	}
}

func TestWorkspaceRepositoryDelete(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := NewWorkspaceRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	email := "workspace-delete-" + uuid.NewString() + "@example.com"

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO users (id, email)
		VALUES ($1, $2)
		`,
		userID,
		email,
	)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)

		if err != nil {
			t.Errorf("cleanup test user: %v", err)
		}
	})

	now := time.Now()

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO workspaces (
			id,
			owner_id,
			name,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
		`,
		workspaceID,
		userID,
		"Delete Workspace",
		now,
		now,
	)
	if err != nil {
		t.Fatalf("insert test workspace: %v", err)
	}

	err = repo.Delete(ctx, workspaceID)
	if err != nil {
		t.Fatalf("delete workspace: %v", err)
	}

	_, err = repo.FindByID(ctx, workspaceID)

	if err != ErrWorkspaceNotFound {
		t.Fatalf(
			"expected ErrWorkspaceNotFound after deletion, got %v",
			err,
		)
	}
}

func TestWorkspaceRepositoryDeleteNotFound(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := NewWorkspaceRepository(db)

	err = repo.Delete(
		ctx,
		uuid.NewString(),
	)

	if err != ErrWorkspaceNotFound {
		t.Fatalf(
			"expected ErrWorkspaceNotFound, got %v",
			err,
		)
	}
}
