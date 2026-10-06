package repository

import (
	"context"
	"errors"

	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

func setupRepositoryFileRepository(
	t *testing.T,
) (context.Context, RepositoryFileRepository, string, *database.DB) {
	t.Helper()

	ctx, db, _, repositoryID := setupRepositorySnapshotRepository(t)

	snapshotRepo := NewRepositorySnapshotRepository(db)

	snapshot := &domain.RepositorySnapshot{
		ID:           uuid.NewString(),
		RepositoryID: repositoryID,
		CommitSHA:    "repository-file-test-commit",
		Branch:       "main",
		CreatedAt:    time.Now().UTC(),
	}

	if err := snapshotRepo.Create(ctx, snapshot); err != nil {
		t.Fatalf("create repository snapshot: %v", err)
	}

	return ctx, NewRepositoryFileRepository(db), snapshot.ID
}

func TestRepositoryFileRepositoryCreateAndFind(t *testing.T) {
	ctx, repo, snapshotID := setupRepositoryFileRepository(t)

	language := "go"

	file := &domain.RepositoryFile{
		ID:         uuid.NewString(),
		SnapshotID: snapshotID,
		Path:       "internal/service/user.go",
		SizeBytes:  1234,
		Language:   &language,
		Content:    "package service\n\nfunc Example() {}\n",
		CreatedAt:  time.Now().UTC(),
	}

	if err := repo.Create(ctx, file); err != nil {
		t.Fatalf("create repository file: %v", err)
	}

	stored, err := repo.FindByID(ctx, file.ID)
	if err != nil {
		t.Fatalf("find repository file: %v", err)
	}

	if stored.ID != file.ID {
		t.Fatalf("expected ID %q, got %q", file.ID, stored.ID)
	}

	if stored.SnapshotID != file.SnapshotID {
		t.Fatalf(
			"expected snapshot ID %q, got %q",
			file.SnapshotID,
			stored.SnapshotID,
		)
	}

	if stored.Path != file.Path {
		t.Fatalf(
			"expected path %q, got %q",
			file.Path,
			stored.Path,
		)
	}

	if stored.SizeBytes != file.SizeBytes {
		t.Fatalf(
			"expected size %d, got %d",
			file.SizeBytes,
			stored.SizeBytes,
		)
	}

	if stored.Language == nil {
		t.Fatal("expected language to be non-nil")
	}

	if *stored.Language != *file.Language {
		t.Fatalf(
			"expected language %q, got %q",
			*file.Language,
			*stored.Language,
		)
	}

	if stored.Content != file.Content {
		t.Fatalf(
			"expected content %q, got %q",
			file.Content,
			stored.Content,
		)
	}
}

func TestRepositoryFileRepositoryCreateAllowsNilLanguage(
	t *testing.T,
) {
	ctx, repo, snapshotID := setupRepositoryFileRepository(t)

	file := &domain.RepositoryFile{
		ID:         uuid.NewString(),
		SnapshotID: snapshotID,
		Path:       "README.md",
		SizeBytes:  100,
		Language:   nil,
		Content:    "# DevFlow AI\n",
		CreatedAt:  time.Now().UTC(),
	}

	if err := repo.Create(ctx, file); err != nil {
		t.Fatalf("create repository file: %v", err)
	}

	stored, err := repo.FindByID(ctx, file.ID)
	if err != nil {
		t.Fatalf("find repository file: %v", err)
	}

	if stored.Language != nil {
		t.Fatalf(
			"expected language to be nil, got %q",
			*stored.Language,
		)
	}
}

func TestRepositoryFileRepositoryCreateRejectsDuplicatePath(
	t *testing.T,
) {
	ctx, repo, snapshotID := setupRepositoryFileRepository(t)

	first := &domain.RepositoryFile{
		ID:         uuid.NewString(),
		SnapshotID: snapshotID,
		Path:       "main.go",
		SizeBytes:  100,
		Content:    "package main\n",
		CreatedAt:  time.Now().UTC(),
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf(
			"create first repository file: %v",
			err,
		)
	}

	second := &domain.RepositoryFile{
		ID:         uuid.NewString(),
		SnapshotID: snapshotID,
		Path:       "main.go",
		SizeBytes:  200,
		Content:    "package main\n\nfunc main() {}\n",
		CreatedAt:  time.Now().UTC(),
	}

	if err := repo.Create(ctx, second); err == nil {
		t.Fatal(
			"expected duplicate repository file creation to fail",
		)
	}
}

func TestRepositoryFileRepositoryFindNotFound(t *testing.T) {
	ctx, repo, _ := setupRepositoryFileRepository(t)

	_, err := repo.FindByID(ctx, uuid.NewString())
	if !errors.Is(err, ErrRepositoryFileNotFound) {
		t.Fatalf(
			"expected ErrRepositoryFileNotFound, got %v",
			err,
		)
	}
}

func TestRepositoryFileRepositoryListBySnapshotID(
	t *testing.T,
) {
	ctx, repo, snapshotID := setupRepositoryFileRepository(t)

	files := []*domain.RepositoryFile{
		{
			ID:         uuid.NewString(),
			SnapshotID: snapshotID,
			Path:       "z-last.go",
			SizeBytes:  50,
			Content:    "package last\n",
			CreatedAt:  time.Now().UTC(),
		},
		{
			ID:         uuid.NewString(),
			SnapshotID: snapshotID,
			Path:       "a-first.go",
			SizeBytes:  60,
			Content:    "package first\n",
			CreatedAt:  time.Now().UTC(),
		},
		{
			ID:         uuid.NewString(),
			SnapshotID: snapshotID,
			Path:       "internal/service/service.go",
			SizeBytes:  70,
			Content:    "package service\n",
			CreatedAt:  time.Now().UTC(),
		},
	}

	for _, file := range files {
		if err := repo.Create(ctx, file); err != nil {
			t.Fatalf(
				"create repository file %q: %v",
				file.Path,
				err,
			)
		}
	}

	stored, err := repo.ListBySnapshotID(ctx, snapshotID)
	if err != nil {
		t.Fatalf(
			"list repository files: %v",
			err,
		)
	}

	if len(stored) != 3 {
		t.Fatalf(
			"expected 3 repository files, got %d",
			len(stored),
		)
	}

	expectedPaths := []string{
		"a-first.go",
		"internal/service/service.go",
		"z-last.go",
	}

	for i, expectedPath := range expectedPaths {
		if stored[i].Path != expectedPath {
			t.Fatalf(
				"expected file %d to have path %q, got %q",
				i,
				expectedPath,
				stored[i].Path,
			)
		}
	}
}

func TestRepositoryFileRepositoryListBySnapshotIDUnknownSnapshot(
	t *testing.T,
) {
	ctx, repo, _ := setupRepositoryFileRepository(t)

	files, err := repo.ListBySnapshotID(
		ctx,
		uuid.NewString(),
	)
	if err != nil {
		t.Fatalf(
			"list repository files: %v",
			err,
		)
	}

	if len(files) != 0 {
		t.Fatalf(
			"expected empty result, got %d files",
			len(files),
		)
	}
}

func TestRepositoryFileRepositorySnapshotDeleteCascadesFiles(
	t *testing.T,
) {
	ctx, repo, snapshotID := setupRepositoryFileRepository(t)

	file := &domain.RepositoryFile{
		ID:         uuid.NewString(),
		SnapshotID: snapshotID,
		Path:       "main.go",
		SizeBytes:  100,
		Content:    "package main\n",
		CreatedAt:  time.Now().UTC(),
	}

	if err := repo.Create(ctx, file); err != nil {
		t.Fatalf(
			"create repository file: %v",
			err,
		)
	}

	// Reuse the existing snapshot test setup to access the
	// database through the repository connection.
	_, db, _, _ := setupRepositorySnapshotRepository(t)

	_, err := db.Pool.Exec(
		ctx,
		`DELETE FROM repository_snapshots WHERE id = $1`,
		snapshotID,
	)
	if err != nil {
		t.Fatalf(
			"delete repository snapshot: %v",
			err,
		)
	}

	_, err = repo.FindByID(ctx, file.ID)
	if !errors.Is(err, ErrRepositoryFileNotFound) {
		t.Fatalf(
			"expected ErrRepositoryFileNotFound after snapshot deletion, got %v",
			err,
		)
	}
}
