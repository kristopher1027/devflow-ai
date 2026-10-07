package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/integration/github"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type ingestProjectRepository struct {
	project *domain.Project
}

func (f *ingestProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	return nil
}

func (f *ingestProjectRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	return f.project, nil
}

func (f *ingestProjectRepository) ListByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
	return nil, nil
}

func (f *ingestProjectRepository) Delete(ctx context.Context, id string) error {
	return nil
}

type ingestConnectionRepository struct {
	connection *domain.GitHubConnection
}

func (f *ingestConnectionRepository) Create(ctx context.Context, c *domain.GitHubConnection) error {
	return nil
}

func (f *ingestConnectionRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) (*domain.GitHubConnection, error) {
	return f.connection, nil
}

func (f *ingestConnectionRepository) FindByID(ctx context.Context, id string) (*domain.GitHubConnection, error) {
	return f.connection, nil
}

func (f *ingestConnectionRepository) FindByInstallationID(ctx context.Context, installationID string) (*domain.GitHubConnection, error) {
	return f.connection, nil
}

func (f *ingestConnectionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return nil
}

type ingestFileRepository struct {
	existing []*domain.RepositoryFile
	created  []*domain.RepositoryFile
}

func (f *ingestFileRepository) Create(ctx context.Context, file *domain.RepositoryFile) error {
	f.created = append(f.created, file)
	return nil
}

func (f *ingestFileRepository) FindByID(ctx context.Context, id string) (*domain.RepositoryFile, error) {
	return nil, repository.ErrRepositoryFileNotFound
}

func (f *ingestFileRepository) ListBySnapshotID(ctx context.Context, snapshotID string) ([]*domain.RepositoryFile, error) {
	return f.existing, nil
}

func (f *ingestFileRepository) createdPaths() []string {
	paths := make([]string, 0, len(f.created))
	for _, file := range f.created {
		paths = append(paths, file.Path)
	}
	return paths
}

type ingestGitHubClient struct {
	tree        *github.RepositoryTree
	treeErr     error
	blobs       map[string]*github.RepositoryBlob
	blobCalls   []string
	installID   string
	owner       string
	name        string
	treeSHAUsed string
}

func (f *ingestGitHubClient) GetLatestCommitSHA(ctx context.Context, installationID, owner, repo, branch string) (string, error) {
	return "", nil
}

func (f *ingestGitHubClient) GetTree(ctx context.Context, installationID, owner, repo, treeSHA string) (*github.RepositoryTree, error) {
	f.installID, f.owner, f.name, f.treeSHAUsed = installationID, owner, repo, treeSHA
	return f.tree, f.treeErr
}

func (f *ingestGitHubClient) GetBlob(ctx context.Context, installationID, owner, repo, blobSHA string) (*github.RepositoryBlob, error) {
	f.blobCalls = append(f.blobCalls, blobSHA)
	return f.blobs[blobSHA], nil
}

func base64Blob(text string) *github.RepositoryBlob {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	// GitHub wraps base64 content with newlines.
	encoded = encoded[:4] + "\n" + encoded[4:]
	return &github.RepositoryBlob{Content: encoded, Encoding: "base64"}
}

func newIngestService(
	snapshotRepo *fakeRepositorySnapshotRepository,
	files *ingestFileRepository,
	client *ingestGitHubClient,
) RepositoryIngestionService {
	return NewRepositoryIngestionService(
		&fakeRepositorySyncRepository{repository: &domain.Repository{
			ID: "repo-1", ProjectID: "project-1", Owner: "octo", Name: "demo",
		}},
		&ingestProjectRepository{project: &domain.Project{ID: "project-1", WorkspaceID: "ws-1"}},
		&ingestConnectionRepository{connection: &domain.GitHubConnection{InstallationID: "inst-9"}},
		snapshotRepo,
		files,
		client,
	)
}

func TestRepositoryIngestionStoresOnlyUsefulTextFiles(t *testing.T) {
	client := &ingestGitHubClient{
		tree: &github.RepositoryTree{SHA: "tree", Entries: []github.RepositoryTreeEntry{
			{Path: "main.go", Type: "blob", SHA: "sha-main", Size: 20},
			{Path: "docs/README.md", Type: "blob", SHA: "sha-readme", Size: 20},
			{Path: "src", Type: "tree", SHA: "sha-dir"},
			{Path: "node_modules/lib/index.js", Type: "blob", SHA: "sha-nm", Size: 20},
			{Path: "package-lock.json", Type: "blob", SHA: "sha-lock", Size: 20},
			{Path: "logo.png", Type: "blob", SHA: "sha-png", Size: 20},
			{Path: "huge.go", Type: "blob", SHA: "sha-huge", Size: maxIngestFileBytes + 1},
			{Path: "binary.go", Type: "blob", SHA: "sha-bin", Size: 20},
		}},
		blobs: map[string]*github.RepositoryBlob{
			"sha-main":   base64Blob("package main\n"),
			"sha-readme": {Content: "# Demo\n", Encoding: "utf-8"},
			"sha-bin":    base64Blob("ab\x00cd"),
		},
	}
	files := &ingestFileRepository{}
	snapshots := &fakeRepositorySnapshotRepository{snapshot: &domain.RepositorySnapshot{
		ID: "snap-1", RepositoryID: "repo-1", CommitSHA: "commit-abc",
	}}

	err := newIngestService(snapshots, files, client).Ingest(context.Background(), "snap-1")

	require.NoError(t, err)
	require.Equal(t, []string{"main.go", "docs/README.md"}, files.createdPaths())
	require.Equal(t, "package main\n", files.created[0].Content)
	require.Equal(t, "go", *files.created[0].Language)
	require.Equal(t, "markdown", *files.created[1].Language)
	require.Equal(t, "snap-1", files.created[0].SnapshotID)
	require.Equal(t, int64(len("package main\n")), files.created[0].SizeBytes)
	require.Equal(t, "inst-9", client.installID)
	require.Equal(t, "octo", client.owner)
	require.Equal(t, "demo", client.name)
	require.Equal(t, "commit-abc", client.treeSHAUsed)
	require.NotContains(t, client.blobCalls, "sha-nm")
	require.NotContains(t, client.blobCalls, "sha-lock")
	require.NotContains(t, client.blobCalls, "sha-huge")
}

func TestRepositoryIngestionResumesAfterPartialFailure(t *testing.T) {
	client := &ingestGitHubClient{
		tree: &github.RepositoryTree{Entries: []github.RepositoryTreeEntry{
			{Path: "a.go", Type: "blob", SHA: "sha-a", Size: 10},
			{Path: "b.go", Type: "blob", SHA: "sha-b", Size: 10},
		}},
		blobs: map[string]*github.RepositoryBlob{
			"sha-a": base64Blob("package a"),
			"sha-b": base64Blob("package b"),
		},
	}
	files := &ingestFileRepository{
		existing: []*domain.RepositoryFile{{Path: "a.go"}},
	}
	snapshots := &fakeRepositorySnapshotRepository{snapshot: &domain.RepositorySnapshot{
		ID: "snap-1", RepositoryID: "repo-1", CommitSHA: "c",
	}}

	require.NoError(t, newIngestService(snapshots, files, client).Ingest(context.Background(), "snap-1"))

	require.Equal(t, []string{"b.go"}, files.createdPaths())
	require.Equal(t, []string{"sha-b"}, client.blobCalls)
}

func TestRepositoryIngestionSnapshotNotFound(t *testing.T) {
	snapshots := &fakeRepositorySnapshotRepository{findErr: repository.ErrRepositorySnapshotNotFound}

	err := newIngestService(snapshots, &ingestFileRepository{}, &ingestGitHubClient{}).
		Ingest(context.Background(), "missing")

	require.ErrorIs(t, err, ErrRepositoryIngestionSnapshotNotFound)
}

func TestRepositoryIngestionTreeErrorStoresNothing(t *testing.T) {
	client := &ingestGitHubClient{treeErr: errors.New("github down")}
	files := &ingestFileRepository{}
	snapshots := &fakeRepositorySnapshotRepository{snapshot: &domain.RepositorySnapshot{
		ID: "snap-1", RepositoryID: "repo-1", CommitSHA: "c",
	}}

	err := newIngestService(snapshots, files, client).Ingest(context.Background(), "snap-1")

	require.EqualError(t, err, "github down")
	require.Empty(t, files.created)
}

func TestSelectIngestCandidatesCapsAndPrefersShallowFiles(t *testing.T) {
	entries := make([]github.RepositoryTreeEntry, 0, maxIngestFiles+1)
	for i := 0; i < maxIngestFiles; i++ {
		entries = append(entries, github.RepositoryTreeEntry{
			Path: fmt.Sprintf("deep/er/est/file%d.go", i), Type: "blob", SHA: "s", Size: 5,
		})
	}
	entries = append(entries, github.RepositoryTreeEntry{
		Path: "main.go", Type: "blob", SHA: "s", Size: 5,
	})

	candidates := selectIngestCandidates(entries)

	require.Len(t, candidates, maxIngestFiles)
	require.Equal(t, "main.go", candidates[0].entry.Path)
}

func TestDecodeTextBlobRejectsBadInput(t *testing.T) {
	for name, blob := range map[string]*github.RepositoryBlob{
		"nil":            nil,
		"empty":          {Content: "", Encoding: "utf-8"},
		"invalid base64": {Content: "!!!not-base64!!!", Encoding: "base64"},
		"invalid utf8":   {Content: string([]byte{0xff, 0xfe}), Encoding: "utf-8"},
		"nul byte":       {Content: "a\x00b", Encoding: "utf-8"},
	} {
		_, ok := decodeTextBlob(blob)
		require.False(t, ok, name)
	}

	text, ok := decodeTextBlob(&github.RepositoryBlob{Content: "hi", Encoding: "utf-8"})
	require.True(t, ok)
	require.Equal(t, "hi", text)
	require.True(t, strings.HasPrefix(text, "h"))
}


func TestRepositoryIngestionRejectsTruncatedTree(t *testing.T) {
	client := &ingestGitHubClient{
		tree: &github.RepositoryTree{
			SHA:       "tree",
			Truncated: true,
			Entries: []github.RepositoryTreeEntry{
				{
					Path: "main.go",
					Type: "blob",
					SHA:  "sha-main",
					Size: 20,
				},
			},
		},
	}

	files := &ingestFileRepository{}

	snapshots := &fakeRepositorySnapshotRepository{
		snapshot: &domain.RepositorySnapshot{
			ID:          "snap-1",
			RepositoryID: "repo-1",
			CommitSHA:   "commit-abc",
		},
	}

	err := newIngestService(
		snapshots,
		files,
		client,
	).Ingest(context.Background(), "snap-1")

	require.ErrorIs(
		t,
		err,
		ErrRepositoryIngestionTreeTruncated,
	)

	require.Empty(t, files.created)
	require.Empty(t, client.blobCalls)
}
