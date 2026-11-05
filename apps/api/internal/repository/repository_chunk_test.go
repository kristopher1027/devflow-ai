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

func TestRepositoryChunkRepositorySearchRanksByMatchedTermsBeforeLimit(t *testing.T) {
	ctx, repo, existingFileID, db := setupRepositoryChunkRepository(t)
	var snapshotID string
	if err := db.Pool.QueryRow(ctx, "SELECT snapshot_id FROM repository_files WHERE id = $1", existingFileID).Scan(&snapshotID); err != nil {
		t.Fatalf("load snapshot ID: %v", err)
	}
	fileRepo := NewRepositoryFileRepository(db)
	now := time.Now().UTC()

	fixtures := []struct {
		path    string
		content string
	}{
		{path: "a_low_relevance.go", content: "repository is mentioned here"},
		{path: "z_high_relevance.go", content: "repository snapshot chunk retrieval"},
	}

	for _, fixture := range fixtures {
		file := &domain.RepositoryFile{
			ID: uuid.NewString(), SnapshotID: snapshotID,
			Path: fixture.path, SizeBytes: int64(len(fixture.content)),
			Content: fixture.content, CreatedAt: now,
		}
		if err := fileRepo.Create(ctx, file); err != nil {
			t.Fatalf("create repository file %q: %v", fixture.path, err)
		}
		chunk := &domain.RepositoryChunk{
			ID: uuid.NewString(), FileID: file.ID, ChunkIndex: 0,
			StartLine: 1, EndLine: 1, CharacterCount: len(fixture.content),
			Content: fixture.content, CreatedAt: now,
		}
		if err := repo.Create(ctx, chunk); err != nil {
			t.Fatalf("create chunk for %q: %v", fixture.path, err)
		}
	}

	results, err := repo.SearchBySnapshotID(ctx, snapshotID, "repository snapshot chunk", 1)
	if err != nil {
		t.Fatalf("search repository chunks: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 result due to limit, got %d", len(results))
	}
	if results[0].FilePath != "z_high_relevance.go" {
		t.Fatalf("expected highest-relevance file before limit, got %q", results[0].FilePath)
	}
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


func TestRepositoryChunkRepositorySearchIsolatesSnapshots(t *testing.T) {
	ctx, repo, existingFileID, db := setupRepositoryChunkRepository(t)

	var repositoryID string
	if err := db.Pool.QueryRow(ctx, `
		SELECT s.repository_id
		FROM repository_snapshots s
		INNER JOIN repository_files f ON f.snapshot_id = s.id
		WHERE f.id = $1
	`, existingFileID).Scan(&repositoryID); err != nil {
		t.Fatalf("load repository ID: %v", err)
	}

	var firstSnapshotID string
	if err := db.Pool.QueryRow(ctx, `
		SELECT snapshot_id FROM repository_files WHERE id = $1
	`, existingFileID).Scan(&firstSnapshotID); err != nil {
		t.Fatalf("load first snapshot ID: %v", err)
	}

	secondSnapshot := &domain.RepositorySnapshot{
		ID:           uuid.NewString(),
		RepositoryID: repositoryID,
		CommitSHA:    "snapshot-isolation-test-" + uuid.NewString(),
		Branch:       "main",
		CreatedAt:    time.Now().UTC(),
	}
	snapshotRepo := NewRepositorySnapshotRepository(db)
	if err := snapshotRepo.Create(ctx, secondSnapshot); err != nil {
		t.Fatalf("create second snapshot: %v", err)
	}

	fileRepo := NewRepositoryFileRepository(db)
	now := time.Now().UTC()
	fixtures := []struct {
		snapshotID string
		path       string
		content    string
	}{
		{firstSnapshotID, "first.go", "snapshot isolation needle"},
		{secondSnapshot.ID, "second.go", "snapshot isolation needle"},
	}

	for _, fixture := range fixtures {
		file := &domain.RepositoryFile{
			ID:         uuid.NewString(),
			SnapshotID: fixture.snapshotID,
			Path:       fixture.path,
			SizeBytes:  int64(len(fixture.content)),
			Content:    fixture.content,
			CreatedAt:  now,
		}
		if err := fileRepo.Create(ctx, file); err != nil {
			t.Fatalf("create file %q: %v", fixture.path, err)
		}

		chunk := &domain.RepositoryChunk{
			ID:             uuid.NewString(),
			FileID:         file.ID,
			ChunkIndex:     0,
			StartLine:      1,
			EndLine:        1,
			CharacterCount: len(fixture.content),
			Content:        fixture.content,
			CreatedAt:      now,
		}
		if err := repo.Create(ctx, chunk); err != nil {
			t.Fatalf("create chunk for %q: %v", fixture.path, err)
		}
	}

	results, err := repo.SearchBySnapshotID(
		ctx, firstSnapshotID, "snapshot isolation needle", 10,
	)
	if err != nil {
		t.Fatalf("search first snapshot: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result from first snapshot, got %d", len(results))
	}
	if results[0].FilePath != "first.go" {
		t.Fatalf("expected first.go, got %q", results[0].FilePath)
	}
}

func TestRepositoryChunkRepositorySearchUsesDeterministicTieBreaking(t *testing.T) {
	ctx, repo, existingFileID, db := setupRepositoryChunkRepository(t)

	var snapshotID string
	if err := db.Pool.QueryRow(ctx, `
		SELECT snapshot_id FROM repository_files WHERE id = $1
	`, existingFileID).Scan(&snapshotID); err != nil {
		t.Fatalf("load snapshot ID: %v", err)
	}

	fileRepo := NewRepositoryFileRepository(db)
	now := time.Now().UTC()
	files := make(map[string]string)

	fixtures := []struct {
		path       string
		chunkIndex int
	}{
		{path: "z-last.go", chunkIndex: 0},
		{path: "a-first.go", chunkIndex: 1},
		{path: "a-first.go", chunkIndex: 0},
	}

	for _, fixture := range fixtures {
		fileID, exists := files[fixture.path]
		if !exists {
			file := &domain.RepositoryFile{
				ID:         uuid.NewString(),
				SnapshotID: snapshotID,
				Path:       fixture.path,
				SizeBytes:  int64(len("retrieval needle")),
				Content:    "retrieval needle",
				CreatedAt:  now,
			}
			if err := fileRepo.Create(ctx, file); err != nil {
				t.Fatalf("create file %q: %v", fixture.path, err)
			}
			fileID = file.ID
			files[fixture.path] = fileID
		}

		chunk := &domain.RepositoryChunk{
			ID:             uuid.NewString(),
			FileID:         fileID,
			ChunkIndex:     fixture.chunkIndex,
			StartLine:      1,
			EndLine:        1,
			CharacterCount: len("retrieval needle"),
			Content:        "retrieval needle",
			CreatedAt:      now,
		}
		if err := repo.Create(ctx, chunk); err != nil {
			t.Fatalf("create chunk %q index %d: %v",
				fixture.path, fixture.chunkIndex, err)
		}
	}

	results, err := repo.SearchBySnapshotID(ctx, snapshotID, "retrieval needle", 10)
	if err != nil {
		t.Fatalf("search repository chunks: %v", err)
	}

	expected := []struct {
		path       string
		chunkIndex int
	}{
		{path: "a-first.go", chunkIndex: 0},
		{path: "a-first.go", chunkIndex: 1},
		{path: "z-last.go", chunkIndex: 0},
	}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(results))
	}

	for i, want := range expected {
		if results[i].FilePath != want.path ||
			results[i].Chunk.ChunkIndex != want.chunkIndex {
			t.Fatalf(
				"result %d: expected %s chunk %d, got %s chunk %d",
				i, want.path, want.chunkIndex,
				results[i].FilePath, results[i].Chunk.ChunkIndex,
			)
		}
	}
}
