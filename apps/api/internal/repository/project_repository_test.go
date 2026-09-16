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

func TestProjectRepositoryCreate(t *testing.T) {
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

	repo := NewProjectRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()

	email := "project-create-" + uuid.NewString() + "@example.com"

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

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO workspaces (
			id,
			owner_id,
			name
		)
		VALUES ($1, $2, $3)
		`,
		workspaceID,
		userID,
		"Project Test Workspace",
	)
	if err != nil {
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
	})

	description := "Test project description"
	now := time.Now()

	project := &domain.Project{
		ID:          projectID,
		WorkspaceID: workspaceID,
		Name:        "Test Project",
		Description: &description,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = repo.Create(ctx, project)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	var (
		storedWorkspaceID string
		storedName        string
		storedDescription string
		storedCreatedBy   string
	)

	err = db.Pool.QueryRow(
		ctx,
		`
		SELECT
			workspace_id,
			name,
			description,
			created_by
		FROM projects
		WHERE id = $1
		`,
		projectID,
	).Scan(
		&storedWorkspaceID,
		&storedName,
		&storedDescription,
		&storedCreatedBy,
	)
	if err != nil {
		t.Fatalf("query created project: %v", err)
	}

	if storedWorkspaceID != workspaceID {
		t.Fatalf(
			"expected workspace ID %s, got %s",
			workspaceID,
			storedWorkspaceID,
		)
	}

	if storedName != project.Name {
		t.Fatalf(
			"expected project name %s, got %s",
			project.Name,
			storedName,
		)
	}

	if storedDescription != description {
		t.Fatalf(
			"expected description %s, got %s",
			description,
			storedDescription,
		)
	}

	if storedCreatedBy != userID {
		t.Fatalf(
			"expected created by %s, got %s",
			userID,
			storedCreatedBy,
		)
	}

	_, err = db.Pool.Exec(
		ctx,
		"DELETE FROM projects WHERE id = $1",
		projectID,
	)
	if err != nil {
		t.Fatalf("cleanup test project: %v", err)
	}
}

func TestProjectRepositoryFindByID(t *testing.T) {
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

	repo := NewProjectRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()

	email := "project-find-" + uuid.NewString() + "@example.com"

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

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO workspaces (
			id,
			owner_id,
			name
		)
		VALUES ($1, $2, $3)
		`,
		workspaceID,
		userID,
		"Find Project Workspace",
	)
	if err != nil {
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
	})

	description := "Find project description"
	now := time.Now()

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO projects (
			id,
			workspace_id,
			name,
			description,
			created_by,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		projectID,
		workspaceID,
		"Find Project",
		description,
		userID,
		now,
		now,
	)
	if err != nil {
		t.Fatalf("insert test project: %v", err)
	}

	project, err := repo.FindByID(ctx, projectID)
	if err != nil {
		t.Fatalf("find project: %v", err)
	}

	if project.ID != projectID {
		t.Fatalf(
			"expected project ID %s, got %s",
			projectID,
			project.ID,
		)
	}

	if project.WorkspaceID != workspaceID {
		t.Fatalf(
			"expected workspace ID %s, got %s",
			workspaceID,
			project.WorkspaceID,
		)
	}

	if project.Name != "Find Project" {
		t.Fatalf(
			"expected project name %q, got %q",
			"Find Project",
			project.Name,
		)
	}

	if project.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *project.Description != description {
		t.Fatalf(
			"expected description %q, got %q",
			description,
			*project.Description,
		)
	}

	if project.CreatedBy != userID {
		t.Fatalf(
			"expected created by %s, got %s",
			userID,
			project.CreatedBy,
		)
	}
}

func TestProjectRepositoryFindByIDNotFound(t *testing.T) {
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

	repo := NewProjectRepository(db)

	_, err = repo.FindByID(ctx, uuid.NewString())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrProjectNotFound {
		t.Fatalf(
			"expected ErrProjectNotFound, got %v",
			err,
		)
	}
}

func TestProjectRepositoryListByWorkspaceID(t *testing.T) {
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

	repo := NewProjectRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()

	email := "project-list-" + uuid.NewString() + "@example.com"

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

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO workspaces (
			id,
			owner_id,
			name
		)
		VALUES ($1, $2, $3)
		`,
		workspaceID,
		userID,
		"List Project Workspace",
	)
	if err != nil {
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
	})

	now := time.Now()

	project1ID := uuid.NewString()
	project2ID := uuid.NewString()

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO projects (
			id,
			workspace_id,
			name,
			created_by,
			created_at,
			updated_at
		)
		VALUES
			($1, $3, $4, $2, $5, $5),
			($6, $3, $7, $2, $8, $8)
		`,
		project1ID,
		userID,
		workspaceID,
		"Project One",
		now,
		project2ID,
		"Project Two",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("insert test projects: %v", err)
	}

	projects, err := repo.ListByWorkspaceID(ctx, workspaceID)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf(
			"expected 2 projects, got %d",
			len(projects),
		)
	}

	if projects[0].Name != "Project One" {
		t.Fatalf(
			"expected first project %q, got %q",
			"Project One",
			projects[0].Name,
		)
	}

	if projects[1].Name != "Project Two" {
		t.Fatalf(
			"expected second project %q, got %q",
			"Project Two",
			projects[1].Name,
		)
	}
}

func TestProjectRepositoryDelete(t *testing.T) {
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

	repo := NewProjectRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()

	email := "project-delete-" + uuid.NewString() + "@example.com"

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

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO workspaces (
			id,
			owner_id,
			name
		)
		VALUES ($1, $2, $3)
		`,
		workspaceID,
		userID,
		"Delete Project Workspace",
	)
	if err != nil {
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
	})

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO projects (
			id,
			workspace_id,
			name,
			created_by
		)
		VALUES ($1, $2, $3, $4)
		`,
		projectID,
		workspaceID,
		"Delete Project",
		userID,
	)
	if err != nil {
		t.Fatalf("insert test project: %v", err)
	}

	err = repo.Delete(ctx, projectID)
	if err != nil {
		t.Fatalf("delete project: %v", err)
	}

	_, err = repo.FindByID(ctx, projectID)
	if err != ErrProjectNotFound {
		t.Fatalf(
			"expected ErrProjectNotFound after delete, got %v",
			err,
		)
	}
}

func TestProjectRepositoryDeleteNotFound(t *testing.T) {
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

	repo := NewProjectRepository(db)

	err = repo.Delete(ctx, uuid.NewString())
	if err != ErrProjectNotFound {
		t.Fatalf(
			"expected ErrProjectNotFound, got %v",
			err,
		)
	}
}
