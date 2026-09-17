package migrations_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/database"
)

func TestGitHubConnectionsMigrationConstraints(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	defer db.Close()

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(
		ctx,
		`SELECT 1 FROM information_schema.tables WHERE table_name = 'github_connections'`,
	); err != nil {
		t.Fatalf("check github connections table: %v", err)
	}

	var tableExists bool
	if err := tx.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'github_connections'
		)`,
	).Scan(&tableExists); err != nil {
		t.Fatalf("check github connections table: %v", err)
	}

	if !tableExists {
		migrationPath := filepath.Join("000008_create_github_connections.up.sql")
		migrationSQL, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read migration: %v", err)
		}

		if _, err := tx.Exec(ctx, string(migrationSQL)); err != nil {
			t.Fatalf("apply migration: %v", err)
		}
	}

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	secondWorkspaceID := uuid.NewString()
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO users (id, email) VALUES ($1, $2)`,
		userID,
		userID+"@example.com",
	); err != nil {
		t.Fatalf("create test user: %v", err)
	}

	for _, id := range []string{workspaceID, secondWorkspaceID} {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO workspaces (id, owner_id, name) VALUES ($1, $2, $3)`,
			id,
			userID,
			"GitHub Test Workspace "+id,
		); err != nil {
			t.Fatalf("create test workspace: %v", err)
		}
	}

	installationID := "installation-" + uuid.NewString()
	insertConnection := func(workspaceID string, installationID string, status string) error {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO github_connections (
				workspace_id, installation_id, account_login, status
			) VALUES ($1, $2, $3, $4)`,
			workspaceID,
			installationID,
			"devflow-org",
			status,
		)
		return err
	}

	if err := insertConnection(
		workspaceID,
		installationID,
		"active",
	); err != nil {
		t.Fatalf("insert valid github connection: %v", err)
	}

	if _, err := tx.Exec(ctx, "SAVEPOINT duplicate_workspace"); err != nil {
		t.Fatalf("create workspace savepoint: %v", err)
	}
	if err := insertConnection(
		workspaceID,
		"installation-"+uuid.NewString(),
		"active",
	); err == nil {
		t.Fatal("expected duplicate workspace connection to fail")
	}
	if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT duplicate_workspace"); err != nil {
		t.Fatalf("rollback workspace savepoint: %v", err)
	}

	if _, err := tx.Exec(ctx, "SAVEPOINT duplicate_installation"); err != nil {
		t.Fatalf("create installation savepoint: %v", err)
	}
	if err := insertConnection(
		secondWorkspaceID,
		installationID,
		"active",
	); err == nil {
		t.Fatal("expected duplicate installation to fail")
	}
	if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT duplicate_installation"); err != nil {
		t.Fatalf("rollback installation savepoint: %v", err)
	}

	if _, err := tx.Exec(ctx, "SAVEPOINT invalid_status"); err != nil {
		t.Fatalf("create status savepoint: %v", err)
	}
	if err := insertConnection(
		secondWorkspaceID,
		"installation-"+uuid.NewString(),
		"unknown",
	); err == nil {
		t.Fatal("expected invalid connection status to fail")
	}
	if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT invalid_status"); err != nil {
		t.Fatalf("rollback status savepoint: %v", err)
	}
}
