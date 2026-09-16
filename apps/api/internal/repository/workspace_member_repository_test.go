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

func setupWorkspaceMemberRepositoryTest(t *testing.T) (
	*database.DB,
	WorkspaceMemberRepository,
	WorkspaceRepository,
	string,
	string,
	string,
) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	workspaceRepository := NewWorkspaceRepository(db)
	memberRepository := NewWorkspaceMemberRepository(db)

	userID := uuid.NewString()
	workspaceID := uuid.NewString()

	ownerID := uuid.NewString()

	// Create users required by the foreign keys.
	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO users (
			id,
			email,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4)
		`,
		userID,
		userID+"@example.com",
		time.Now(),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	_, err = db.Pool.Exec(
		ctx,
		`
		INSERT INTO users (
			id,
			email,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4)
		`,
		ownerID,
		ownerID+"@example.com",
		time.Now(),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("create owner test user: %v", err)
	}

	workspace := &domain.Workspace{
		ID:        workspaceID,
		OwnerID:   ownerID,
		Name:      "Workspace Member Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := workspaceRepository.Create(ctx, workspace); err != nil {
		t.Fatalf("create test workspace: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(
			context.Background(),
			"DELETE FROM workspaces WHERE id = $1",
			workspaceID,
		)

		_, _ = db.Pool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id IN ($1, $2)",
			userID,
			ownerID,
		)
	})

	return db, memberRepository, workspaceRepository, workspaceID, userID, ownerID
}

func TestWorkspaceMemberRepositoryCreate(t *testing.T) {
	_, repository, _, workspaceID, userID, _ :=
		setupWorkspaceMemberRepositoryTest(t)

	ctx := context.Background()

	member := &domain.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        "member",
		CreatedAt:   time.Now(),
	}

	if err := repository.Create(ctx, member); err != nil {
		t.Fatalf("create workspace member: %v", err)
	}

	found, err := repository.Find(
		ctx,
		workspaceID,
		userID,
	)
	if err != nil {
		t.Fatalf("find created workspace member: %v", err)
	}

	if found.WorkspaceID != workspaceID {
		t.Fatalf(
			"expected workspace ID %q, got %q",
			workspaceID,
			found.WorkspaceID,
		)
	}

	if found.UserID != userID {
		t.Fatalf(
			"expected user ID %q, got %q",
			userID,
			found.UserID,
		)
	}

	if found.Role != "member" {
		t.Fatalf(
			"expected role %q, got %q",
			"member",
			found.Role,
		)
	}
}

func TestWorkspaceMemberRepositoryCreateDuplicate(t *testing.T) {
	_, repository, _, workspaceID, userID, _ :=
		setupWorkspaceMemberRepositoryTest(t)

	ctx := context.Background()

	member := &domain.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        "member",
		CreatedAt:   time.Now(),
	}

	if err := repository.Create(ctx, member); err != nil {
		t.Fatalf("create first workspace member: %v", err)
	}

	err := repository.Create(ctx, member)
	if err == nil {
		t.Fatal("expected duplicate workspace member error")
	}
}

func TestWorkspaceMemberRepositoryFindNotFound(t *testing.T) {
	_, repository, _, workspaceID, userID, _ :=
		setupWorkspaceMemberRepositoryTest(t)

	ctx := context.Background()

	_, err := repository.Find(
		ctx,
		workspaceID,
		userID,
	)
	if err == nil {
		t.Fatal("expected workspace member not found error")
	}

	if err != ErrWorkspaceMemberNotFound {
		t.Fatalf(
			"expected ErrWorkspaceMemberNotFound, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberRepositoryListByWorkspaceID(t *testing.T) {
	_, repository, _, workspaceID, userID, ownerID :=
		setupWorkspaceMemberRepositoryTest(t)

	ctx := context.Background()

	now := time.Now()

	members := []*domain.WorkspaceMember{
		{
			WorkspaceID: workspaceID,
			UserID:      ownerID,
			Role:        "owner",
			CreatedAt:   now,
		},
		{
			WorkspaceID: workspaceID,
			UserID:      userID,
			Role:        "member",
			CreatedAt:   now.Add(time.Second),
		},
	}

	for _, member := range members {
		if err := repository.Create(ctx, member); err != nil {
			t.Fatalf("create workspace member: %v", err)
		}
	}

	found, err := repository.ListByWorkspaceID(
		ctx,
		workspaceID,
	)
	if err != nil {
		t.Fatalf("list workspace members: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf(
			"expected 2 members, got %d",
			len(found),
		)
	}

	if found[0].UserID != ownerID {
		t.Fatalf(
			"expected first member to be owner %q, got %q",
			ownerID,
			found[0].UserID,
		)
	}

	if found[1].UserID != userID {
		t.Fatalf(
			"expected second member to be %q, got %q",
			userID,
			found[1].UserID,
		)
	}
}

func TestWorkspaceMemberRepositoryUpdateRole(t *testing.T) {
	_, repository, _, workspaceID, userID, _ :=
		setupWorkspaceMemberRepositoryTest(t)

	ctx := context.Background()

	member := &domain.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        "member",
		CreatedAt:   time.Now(),
	}

	if err := repository.Create(ctx, member); err != nil {
		t.Fatalf("create workspace member: %v", err)
	}

	if err := repository.UpdateRole(
		ctx,
		workspaceID,
		userID,
		"admin",
	); err != nil {
		t.Fatalf("update workspace member role: %v", err)
	}

	found, err := repository.Find(
		ctx,
		workspaceID,
		userID,
	)
	if err != nil {
		t.Fatalf("find updated workspace member: %v", err)
	}

	if found.Role != "admin" {
		t.Fatalf(
			"expected role %q, got %q",
			"admin",
			found.Role,
		)
	}
}

func TestWorkspaceMemberRepositoryUpdateRoleNotFound(t *testing.T) {
	_, repository, _, workspaceID, userID, _ :=
		setupWorkspaceMemberRepositoryTest(t)

	ctx := context.Background()

	err := repository.UpdateRole(
		ctx,
		workspaceID,
		userID,
		"admin",
	)
	if err != ErrWorkspaceMemberNotFound {
		t.Fatalf(
			"expected ErrWorkspaceMemberNotFound, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberRepositoryDelete(t *testing.T) {
	_, repository, _, workspaceID, userID, _ :=
		setupWorkspaceMemberRepositoryTest(t)

	ctx := context.Background()

	member := &domain.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        "member",
		CreatedAt:   time.Now(),
	}

	if err := repository.Create(ctx, member); err != nil {
		t.Fatalf("create workspace member: %v", err)
	}

	if err := repository.Delete(
		ctx,
		workspaceID,
		userID,
	); err != nil {
		t.Fatalf("delete workspace member: %v", err)
	}

	_, err := repository.Find(
		ctx,
		workspaceID,
		userID,
	)

	if err != ErrWorkspaceMemberNotFound {
		t.Fatalf(
			"expected ErrWorkspaceMemberNotFound after delete, got %v",
			err,
		)
	}
}

func TestWorkspaceMemberRepositoryDeleteNotFound(t *testing.T) {
	_, repository, _, workspaceID, userID, _ :=
		setupWorkspaceMemberRepositoryTest(t)

	ctx := context.Background()

	err := repository.Delete(
		ctx,
		workspaceID,
		userID,
	)

	if err != ErrWorkspaceMemberNotFound {
		t.Fatalf(
			"expected ErrWorkspaceMemberNotFound, got %v",
			err,
		)
	}
}
