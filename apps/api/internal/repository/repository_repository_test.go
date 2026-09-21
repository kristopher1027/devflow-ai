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

func TestRepositoryRepositoryCreate(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()
	repositoryID := uuid.NewString()

	email := "repository-create-" + uuid.NewString() + "@example.com"

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
		"Repository Test Workspace",
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
		"Repository Test Project",
		userID,
	)
	if err != nil {
		t.Fatalf("insert test project: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM projects WHERE id = $1",
			projectID,
		)
		if err != nil {
			t.Errorf("cleanup test project: %v", err)
		}
	})

	now := time.Now()

	repository := &domain.Repository{
		ID:            repositoryID,
		ProjectID:     projectID,
		Provider:      "github",
		ExternalID:    "123456789",
		Owner:         "kristopher1027",
		Name:          "devflow-ai",
		FullName:      "kristopher1027/devflow-ai",
		DefaultBranch: "main",
		HTMLURL:       "https://github.com/kristopher1027/devflow-ai",
		CloneURL:      "https://github.com/kristopher1027/devflow-ai.git",
		IsPrivate:     true,
		SyncStatus:    "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err = repo.Create(ctx, repository)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	var (
		storedProjectID     string
		storedProvider      string
		storedExternalID    string
		storedOwner         string
		storedName          string
		storedFullName      string
		storedDefaultBranch string
		storedHTMLURL       string
		storedCloneURL      string
		storedIsPrivate     bool
		storedSyncStatus    string
	)

	err = db.Pool.QueryRow(
		ctx,
		`
		SELECT
			project_id,
			provider,
			external_id,
			owner,
			name,
			full_name,
			default_branch,
			html_url,
			clone_url,
			is_private,
			sync_status
		FROM repositories
		WHERE id = $1
		`,
		repositoryID,
	).Scan(
		&storedProjectID,
		&storedProvider,
		&storedExternalID,
		&storedOwner,
		&storedName,
		&storedFullName,
		&storedDefaultBranch,
		&storedHTMLURL,
		&storedCloneURL,
		&storedIsPrivate,
		&storedSyncStatus,
	)
	if err != nil {
		t.Fatalf("query created repository: %v", err)
	}

	if storedProjectID != projectID {
		t.Fatalf(
			"expected project ID %s, got %s",
			projectID,
			storedProjectID,
		)
	}

	if storedProvider != repository.Provider {
		t.Fatalf(
			"expected provider %s, got %s",
			repository.Provider,
			storedProvider,
		)
	}

	if storedExternalID != repository.ExternalID {
		t.Fatalf(
			"expected external ID %s, got %s",
			repository.ExternalID,
			storedExternalID,
		)
	}

	if storedOwner != repository.Owner {
		t.Fatalf(
			"expected owner %s, got %s",
			repository.Owner,
			storedOwner,
		)
	}

	if storedName != repository.Name {
		t.Fatalf(
			"expected name %s, got %s",
			repository.Name,
			storedName,
		)
	}

	if storedFullName != repository.FullName {
		t.Fatalf(
			"expected full name %s, got %s",
			repository.FullName,
			storedFullName,
		)
	}

	if storedDefaultBranch != repository.DefaultBranch {
		t.Fatalf(
			"expected default branch %s, got %s",
			repository.DefaultBranch,
			storedDefaultBranch,
		)
	}

	if storedHTMLURL != repository.HTMLURL {
		t.Fatalf(
			"expected HTML URL %s, got %s",
			repository.HTMLURL,
			storedHTMLURL,
		)
	}

	if storedCloneURL != repository.CloneURL {
		t.Fatalf(
			"expected clone URL %s, got %s",
			repository.CloneURL,
			storedCloneURL,
		)
	}

	if storedIsPrivate != repository.IsPrivate {
		t.Fatalf(
			"expected private %v, got %v",
			repository.IsPrivate,
			storedIsPrivate,
		)
	}

	if storedSyncStatus != repository.SyncStatus {
		t.Fatalf(
			"expected sync status %s, got %s",
			repository.SyncStatus,
			storedSyncStatus,
		)
	}
}

func TestRepositoryRepositoryFindByID(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()
	repositoryID := uuid.NewString()

	email := "repository-find-" + uuid.NewString() + "@example.com"

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
		"Find Repository Workspace",
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
		"Find Repository Project",
		userID,
	)
	if err != nil {
		t.Fatalf("insert test project: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM projects WHERE id = $1",
			projectID,
		)
		if err != nil {
			t.Errorf("cleanup test project: %v", err)
		}
	})

	//description := "Repository find test"

	now := time.Now()

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO repositories (
			id,
			project_id,
			provider,
			external_id,
			owner,
			name,
			full_name,
			default_branch,
			html_url,
			clone_url,
			is_private,
			sync_status,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14
		)
		`,
		repositoryID,
		projectID,
		"github",
		"987654321",
		"example-owner",
		"example-repo",
		"example-owner/example-repo",
		"main",
		"https://github.com/example-owner/example-repo",
		"https://github.com/example-owner/example-repo.git",
		false,
		"synced",
		now,
		now,
	)
	if err != nil {
		t.Fatalf("insert test repository: %v", err)
	}

	repository, err := repo.FindByID(ctx, repositoryID)
	if err != nil {
		t.Fatalf("find repository: %v", err)
	}

	if repository.ID != repositoryID {
		t.Fatalf(
			"expected repository ID %s, got %s",
			repositoryID,
			repository.ID,
		)
	}

	if repository.ProjectID != projectID {
		t.Fatalf(
			"expected project ID %s, got %s",
			projectID,
			repository.ProjectID,
		)
	}

	if repository.Provider != "github" {
		t.Fatalf(
			"expected provider %q, got %q",
			"github",
			repository.Provider,
		)
	}

	if repository.ExternalID != "987654321" {
		t.Fatalf(
			"expected external ID %q, got %q",
			"987654321",
			repository.ExternalID,
		)
	}

	if repository.Owner != "example-owner" {
		t.Fatalf(
			"expected owner %q, got %q",
			"example-owner",
			repository.Owner,
		)
	}

	if repository.Name != "example-repo" {
		t.Fatalf(
			"expected name %q, got %q",
			"example-repo",
			repository.Name,
		)
	}

	if repository.FullName != "example-owner/example-repo" {
		t.Fatalf(
			"expected full name %q, got %q",
			"example-owner/example-repo",
			repository.FullName,
		)
	}

	if repository.DefaultBranch != "main" {
		t.Fatalf(
			"expected default branch %q, got %q",
			"main",
			repository.DefaultBranch,
		)
	}

	if repository.IsPrivate {
		t.Fatal("expected repository to be public")
	}

	if repository.SyncStatus != "synced" {
		t.Fatalf(
			"expected sync status %q, got %q",
			"synced",
			repository.SyncStatus,
		)
	}

	// if repository.Description != nil {
	// 	t.Fatalf("unexpected repository description: %v", description)
	// }

	if repository.LastSyncedAt != nil {
		t.Fatal("expected last synced at to be nil")
	}
}

func TestRepositoryRepositoryFindByIDNotFound(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	_, err = repo.FindByID(ctx, uuid.NewString())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf(
			"expected ErrRepositoryNotFound, got %v",
			err,
		)
	}
}

func TestRepositoryRepositoryListByProjectID(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()

	email := "repository-list-" + uuid.NewString() + "@example.com"

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
		"List Repository Workspace",
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
		"List Repository Project",
		userID,
	)
	if err != nil {
		t.Fatalf("insert test project: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM projects WHERE id = $1",
			projectID,
		)
		if err != nil {
			t.Errorf("cleanup test project: %v", err)
		}
	})

	now := time.Now()

	repository1ID := uuid.NewString()
	repository2ID := uuid.NewString()

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO repositories (
			id,
			project_id,
			provider,
			external_id,
			owner,
			name,
			full_name,
			default_branch,
			html_url,
			clone_url,
			created_at,
			updated_at
		)
		VALUES
			($1, $2, 'github', $3, $4, $5, $6, 'main',
			 $7, $8, $9, $9),
			($10, $2, 'github', $11, $4, $12, $13, 'main',
			 $14, $15, $16, $16)
		`,
		repository1ID,
		projectID,
		"111111",
		"example-owner",
		"repository-one",
		"example-owner/repository-one",
		"https://github.com/example-owner/repository-one",
		"https://github.com/example-owner/repository-one.git",
		now,
		repository2ID,
		"222222",
		"repository-two",
		"example-owner/repository-two",
		"https://github.com/example-owner/repository-two",
		"https://github.com/example-owner/repository-two.git",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("insert test repositories: %v", err)
	}

	repositories, err := repo.ListByProjectID(ctx, projectID)
	if err != nil {
		t.Fatalf("list repositories: %v", err)
	}

	if len(repositories) != 2 {
		t.Fatalf(
			"expected 2 repositories, got %d",
			len(repositories),
		)
	}

	if repositories[0].Name != "repository-one" {
		t.Fatalf(
			"expected first repository %q, got %q",
			"repository-one",
			repositories[0].Name,
		)
	}

	if repositories[1].Name != "repository-two" {
		t.Fatalf(
			"expected second repository %q, got %q",
			"repository-two",
			repositories[1].Name,
		)
	}
}

func TestRepositoryRepositoryFindByProviderExternalID(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()
	repositoryID := uuid.NewString()

	email := "repository-external-" + uuid.NewString() + "@example.com"

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
		"External Repository Workspace",
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
		"External Repository Project",
		userID,
	)
	if err != nil {
		t.Fatalf("insert test project: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM projects WHERE id = $1",
			projectID,
		)
		if err != nil {
			t.Errorf("cleanup test project: %v", err)
		}
	})

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO repositories (
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
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		`,
		repositoryID,
		projectID,
		"github",
		"555555",
		"external-owner",
		"external-repo",
		"external-owner/external-repo",
		"main",
		"https://github.com/external-owner/external-repo",
		"https://github.com/external-owner/external-repo.git",
	)
	if err != nil {
		t.Fatalf("insert test repository: %v", err)
	}

	repository, err := repo.FindByProviderExternalID(
		ctx,
		"github",
		"555555",
	)
	if err != nil {
		t.Fatalf("find repository by external ID: %v", err)
	}

	if repository.ID != repositoryID {
		t.Fatalf(
			"expected repository ID %s, got %s",
			repositoryID,
			repository.ID,
		)
	}

	if repository.ProjectID != projectID {
		t.Fatalf(
			"expected project ID %s, got %s",
			projectID,
			repository.ProjectID,
		)
	}
}

func TestRepositoryRepositoryFindByProviderExternalIDNotFound(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	_, err = repo.FindByProviderExternalID(
		ctx,
		"github",
		"non-existent",
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf(
			"expected ErrRepositoryNotFound, got %v",
			err,
		)
	}
}

func TestRepositoryRepositoryDelete(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()
	repositoryID := uuid.NewString()

	email := "repository-delete-" + uuid.NewString() + "@example.com"

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
		"Delete Repository Workspace",
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
		"Delete Repository Project",
		userID,
	)
	if err != nil {
		t.Fatalf("insert test project: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM projects WHERE id = $1",
			projectID,
		)
		if err != nil {
			t.Errorf("cleanup test project: %v", err)
		}
	})

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO repositories (
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
		repositoryID,
		projectID,
		"github",
		"777777",
		"delete-owner",
		"delete-repo",
		"delete-owner/delete-repo",
		"main",
		"https://github.com/delete-owner/delete-repo",
		"https://github.com/delete-owner/delete-repo.git",
	)
	if err != nil {
		t.Fatalf("insert test repository: %v", err)
	}

	err = repo.Delete(ctx, repositoryID)
	if err != nil {
		t.Fatalf("delete repository: %v", err)
	}

	_, err = repo.FindByID(ctx, repositoryID)
	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf(
			"expected ErrRepositoryNotFound after delete, got %v",
			err,
		)
	}
}

func TestRepositoryRepositoryDeleteNotFound(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	err = repo.Delete(ctx, uuid.NewString())
	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf(
			"expected ErrRepositoryNotFound, got %v",
			err,
		)
	}
}
func TestRepositoryRepositoryUpdateSyncStatus(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	projectID := uuid.NewString()
	repositoryID := uuid.NewString()

	email := "repository-sync-status-" + uuid.NewString() + "@example.com"

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
		"Repository Sync Status Workspace",
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
		"Repository Sync Status Project",
		userID,
	)
	if err != nil {
		t.Fatalf("insert test project: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM projects WHERE id = $1",
			projectID,
		)
		if err != nil {
			t.Errorf("cleanup test project: %v", err)
		}
	})

	now := time.Now().UTC()

	testRepository := &domain.Repository{
		ID:            repositoryID,
		ProjectID:     projectID,
		Provider:      "github",
		ExternalID:    uuid.NewString(),
		Owner:         "test-owner",
		Name:          "test-repository",
		FullName:      "test-owner/test-repository",
		DefaultBranch: "main",
		HTMLURL:       "https://github.com/test-owner/test-repository",
		CloneURL:      "https://github.com/test-owner/test-repository.git",
		IsPrivate:     true,
		SyncStatus:    "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err = repo.Create(ctx, testRepository)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}
	syncedAt := time.Now().UTC().Truncate(time.Microsecond)

	err = repo.UpdateSyncStatus(
		ctx,
		repositoryID,
		"synced",
		&syncedAt,
	)
	if err != nil {
		t.Fatalf("update repository sync status: %v", err)
	}

	var (
		storedSyncStatus   string
		storedLastSyncedAt *time.Time
	)

	err = db.Pool.QueryRow(
		ctx,
		`
		SELECT
			sync_status,
			last_synced_at
		FROM repositories
		WHERE id = $1
		`,
		repositoryID,
	).Scan(
		&storedSyncStatus,
		&storedLastSyncedAt,
	)
	if err != nil {
		t.Fatalf("query updated repository: %v", err)
	}

	if storedSyncStatus != "synced" {
		t.Fatalf(
			"expected sync status synced, got %s",
			storedSyncStatus,
		)
	}

	if storedLastSyncedAt == nil {
		t.Fatal("expected last synced at to be set")
	}

	if !storedLastSyncedAt.Equal(syncedAt) {
		t.Fatalf(
			"expected last synced at %v, got %v",
			syncedAt,
			*storedLastSyncedAt,
		)
	}
}
func TestRepositoryRepositoryUpdateSyncStatusNotFound(t *testing.T) {
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

	repo := NewRepositoryRepository(db)

	err = repo.UpdateSyncStatus(
		ctx,
		uuid.NewString(),
		"synced",
		func() *time.Time {
			now := time.Now().UTC().Truncate(time.Microsecond)
			return &now
		}(),
	)
	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf(
			"expected ErrRepositoryNotFound, got %v",
			err,
		)
	}
}
