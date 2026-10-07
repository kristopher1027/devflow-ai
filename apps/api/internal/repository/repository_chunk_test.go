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

func setupRepositoryChunkRepository(t *testing.T) (context.Context, RepositoryChunkRepository, string, *database.DB) {
	t.Helper()
	ctx, _, snapshotID, db := setupRepositoryFileRepository(t)
	fileRepo := NewRepositoryFileRepository(db)
	file := &domain.RepositoryFile{
		ID: uuid.NewString(), SnapshotID: snapshotID,
		Path: "internal/service/example.go", SizeBytes: 100,
		Content: "package service\n\nfunc Example() {}\n", CreatedAt: time.Now().UTC(),
	}
	if err := fileRepo.Create(ctx, file); err != nil {
		t.Fatalf("create repository file: %v", err)
	}
	return ctx, NewRepositoryChunkRepository(db), file.ID, db
}

func TestRepositoryChunkRepositoryCreateAndFind(t *testing.T) {
	ctx, repo, fileID, _ := setupRepositoryChunkRepository(t)
	content := "package service\n\nfunc Example() {}\n"
	chunk := &domain.RepositoryChunk{
		ID: uuid.NewString(), FileID: fileID, ChunkIndex: 0,
		StartLine: 1, EndLine: 3, CharacterCount: len(content),
		Content: content, CreatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, chunk); err != nil {
		t.Fatalf("create repository chunk: %v", err)
	}
	stored, err := repo.FindByID(ctx, chunk.ID)
	if err != nil {
		t.Fatalf("find repository chunk: %v", err)
	}
	if stored.ID != chunk.ID || stored.FileID != chunk.FileID ||
		stored.ChunkIndex != chunk.ChunkIndex ||
		stored.StartLine != chunk.StartLine || stored.EndLine != chunk.EndLine ||
		stored.CharacterCount != chunk.CharacterCount || stored.Content != chunk.Content {
		t.Fatalf("stored chunk does not match created chunk: %+v", stored)
	}
}

func TestRepositoryChunkRepositoryFindNotFound(t *testing.T) {
	ctx, repo, _, _ := setupRepositoryChunkRepository(t)
	_, err := repo.FindByID(ctx, uuid.NewString())
	if !errors.Is(err, ErrRepositoryChunkNotFound) {
		t.Fatalf("expected ErrRepositoryChunkNotFound, got %v", err)
	}
}

func TestRepositoryChunkRepositoryListByFileIDOrdersByChunkIndex(t *testing.T) {
	ctx, repo, fileID, _ := setupRepositoryChunkRepository(t)
	chunks := []*domain.RepositoryChunk{
		{ID: uuid.NewString(), FileID: fileID, ChunkIndex: 2, StartLine: 5, EndLine: 6, CharacterCount: 10, Content: "chunk two", CreatedAt: time.Now().UTC()},
		{ID: uuid.NewString(), FileID: fileID, ChunkIndex: 0, StartLine: 1, EndLine: 2, CharacterCount: 10, Content: "chunk zero", CreatedAt: time.Now().UTC()},
		{ID: uuid.NewString(), FileID: fileID, ChunkIndex: 1, StartLine: 3, EndLine: 4, CharacterCount: 10, Content: "chunk one", CreatedAt: time.Now().UTC()},
	}
	for _, chunk := range chunks {
		if err := repo.Create(ctx, chunk); err != nil {
			t.Fatalf("create repository chunk %d: %v", chunk.ChunkIndex, err)
		}
	}
	stored, err := repo.ListByFileID(ctx, fileID)
	if err != nil {
		t.Fatalf("list repository chunks: %v", err)
	}
	if len(stored) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(stored))
	}
	for i, chunk := range stored {
		if chunk.ChunkIndex != i {
			t.Fatalf("expected chunk index %d at position %d, got %d", i, i, chunk.ChunkIndex)
		}
	}
}

func TestRepositoryChunkRepositoryRejectsDuplicateIndex(t *testing.T) {
	ctx, repo, fileID, _ := setupRepositoryChunkRepository(t)
	first := &domain.RepositoryChunk{
		ID: uuid.NewString(), FileID: fileID, ChunkIndex: 0,
		StartLine: 1, EndLine: 2, CharacterCount: 5,
		Content: "first", CreatedAt: time.Now().UTC(),
	}
	second := &domain.RepositoryChunk{
		ID: uuid.NewString(), FileID: fileID, ChunkIndex: 0,
		StartLine: 3, EndLine: 4, CharacterCount: 6,
		Content: "second", CreatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first chunk: %v", err)
	}
	if err := repo.Create(ctx, second); err == nil {
		t.Fatal("expected duplicate chunk index to fail")
	}
}

func TestRepositoryChunkRepositoryFileDeleteCascadesChunks(t *testing.T) {
	ctx, repo, fileID, db := setupRepositoryChunkRepository(t)
	chunk := &domain.RepositoryChunk{
		ID: uuid.NewString(), FileID: fileID, ChunkIndex: 0,
		StartLine: 1, EndLine: 1, CharacterCount: 5,
		Content: "hello", CreatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, chunk); err != nil {
		t.Fatalf("create repository chunk: %v", err)
	}
	if _, err := db.Pool.Exec(ctx, "DELETE FROM repository_files WHERE id = $1", fileID); err != nil {
		t.Fatalf("delete repository file: %v", err)
	}
	_, err := repo.FindByID(ctx, chunk.ID)
	if !errors.Is(err, ErrRepositoryChunkNotFound) {
		t.Fatalf("expected ErrRepositoryChunkNotFound after file deletion, got %v", err)
	}
}
