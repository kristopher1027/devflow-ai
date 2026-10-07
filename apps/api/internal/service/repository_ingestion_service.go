package service

import (
	"context"
	"encoding/base64"
	"errors"
	"path"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/integration/github"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

const (
	// maxIngestFileBytes skips very large files; they are rarely useful
	// as context and bloat the database.
	maxIngestFileBytes = 200_000

	// maxIngestFiles caps how many files one snapshot stores.
	maxIngestFiles = 300
)

var (
	ErrRepositoryIngestionSnapshotNotFound = errors.New(
		"repository ingestion snapshot not found",
	)
	ErrRepositoryIngestionRepositoryNotFound = errors.New(
		"repository ingestion repository not found",
	)
)

type RepositoryIngestionService interface {
	Ingest(
		ctx context.Context,
		snapshotID string,
	) error
}

type RepositoryIngestionServiceImpl struct {
	repositoryRepo       repository.RepositoryRepository
	projectRepo          repository.ProjectRepository
	githubConnectionRepo repository.GitHubConnectionRepository
	snapshotRepo         repository.RepositorySnapshotRepository
	fileRepo             repository.RepositoryFileRepository
	githubClient         github.RepositoryClient
}

func NewRepositoryIngestionService(
	repositoryRepo repository.RepositoryRepository,
	projectRepo repository.ProjectRepository,
	githubConnectionRepo repository.GitHubConnectionRepository,
	snapshotRepo repository.RepositorySnapshotRepository,
	fileRepo repository.RepositoryFileRepository,
	githubClient github.RepositoryClient,
) RepositoryIngestionService {
	return &RepositoryIngestionServiceImpl{
		repositoryRepo:       repositoryRepo,
		projectRepo:          projectRepo,
		githubConnectionRepo: githubConnectionRepo,
		snapshotRepo:         snapshotRepo,
		fileRepo:             fileRepo,
		githubClient:         githubClient,
	}
}

// Ingest downloads the text files of a snapshot's commit and stores them.
// It is safe to call again after a partial failure: files already stored
// for the snapshot are skipped.
func (s *RepositoryIngestionServiceImpl) Ingest(
	ctx context.Context,
	snapshotID string,
) error {
	snapshot, err := s.snapshotRepo.FindByID(ctx, snapshotID)
	if err != nil {
		if errors.Is(err, repository.ErrRepositorySnapshotNotFound) {
			return ErrRepositoryIngestionSnapshotNotFound
		}
		return err
	}

	repo, err := s.repositoryRepo.FindByID(ctx, snapshot.RepositoryID)
	if err != nil {
		if errors.Is(err, repository.ErrRepositoryNotFound) {
			return ErrRepositoryIngestionRepositoryNotFound
		}
		return err
	}

	project, err := s.projectRepo.FindByID(ctx, repo.ProjectID)
	if err != nil {
		return err
	}

	connection, err := s.githubConnectionRepo.FindByWorkspaceID(
		ctx,
		project.WorkspaceID,
	)
	if err != nil {
		return err
	}

	// GitHub's trees endpoint accepts a commit SHA and resolves its tree.
	tree, err := s.githubClient.GetTree(
		ctx,
		connection.InstallationID,
		repo.Owner,
		repo.Name,
		snapshot.CommitSHA,
	)
	if err != nil {
		return err
	}

	existing, err := s.fileRepo.ListBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return err
	}
	alreadyStored := make(map[string]struct{}, len(existing))
	for _, file := range existing {
		alreadyStored[file.Path] = struct{}{}
	}

	for _, candidate := range selectIngestCandidates(tree.Entries) {
		if _, done := alreadyStored[candidate.entry.Path]; done {
			continue
		}

		blob, err := s.githubClient.GetBlob(
			ctx,
			connection.InstallationID,
			repo.Owner,
			repo.Name,
			candidate.entry.SHA,
		)
		if err != nil {
			return err
		}

		content, ok := decodeTextBlob(blob)
		if !ok {
			continue
		}

		language := candidate.language
		file := &domain.RepositoryFile{
			ID:         uuid.NewString(),
			SnapshotID: snapshot.ID,
			Path:       candidate.entry.Path,
			SizeBytes:  int64(len(content)),
			Language:   &language,
			Content:    content,
			CreatedAt:  time.Now().UTC(),
		}
		if err := s.fileRepo.Create(ctx, file); err != nil {
			return err
		}
	}

	return nil
}

type ingestCandidate struct {
	entry    github.RepositoryTreeEntry
	language string
}

// selectIngestCandidates keeps small source/doc files, drops dependency and
// build folders and lock files, and prefers files closer to the repo root
// when the cap is reached.
func selectIngestCandidates(
	entries []github.RepositoryTreeEntry,
) []ingestCandidate {
	candidates := make([]ingestCandidate, 0, len(entries))

	for _, entry := range entries {
		if entry.Type != "blob" ||
			entry.Size <= 0 ||
			entry.Size > maxIngestFileBytes ||
			isIgnoredIngestPath(entry.Path) {
			continue
		}

		language, ok := ingestLanguage(entry.Path)
		if !ok {
			continue
		}

		candidates = append(candidates, ingestCandidate{
			entry:    entry,
			language: language,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return strings.Count(candidates[i].entry.Path, "/") <
			strings.Count(candidates[j].entry.Path, "/")
	})

	if len(candidates) > maxIngestFiles {
		candidates = candidates[:maxIngestFiles]
	}

	return candidates
}

var ignoredIngestDirs = map[string]struct{}{
	".git": {}, "node_modules": {}, "vendor": {}, "dist": {},
	"build": {}, ".next": {}, "coverage": {}, "__pycache__": {},
}

var ignoredIngestFiles = map[string]struct{}{
	"package-lock.json": {}, "yarn.lock": {}, "pnpm-lock.yaml": {},
	"go.sum": {}, "Cargo.lock": {}, "poetry.lock": {},
}

func isIgnoredIngestPath(p string) bool {
	segments := strings.Split(p, "/")
	for _, dir := range segments[:len(segments)-1] {
		if _, ignored := ignoredIngestDirs[dir]; ignored {
			return true
		}
	}
	_, ignored := ignoredIngestFiles[segments[len(segments)-1]]
	return ignored
}

var ingestExtensionLanguages = map[string]string{
	".go": "go", ".py": "python", ".js": "javascript", ".jsx": "javascript",
	".ts": "typescript", ".tsx": "typescript", ".java": "java",
	".rs": "rust", ".rb": "ruby", ".php": "php", ".c": "c", ".h": "c",
	".cpp": "cpp", ".cs": "csharp", ".sql": "sql", ".sh": "shell",
	".md": "markdown", ".json": "json", ".yaml": "yaml", ".yml": "yaml",
	".toml": "toml", ".html": "html", ".css": "css",
}

var ingestFilenameLanguages = map[string]string{
	"Dockerfile": "dockerfile", "Makefile": "makefile",
}

func ingestLanguage(p string) (string, bool) {
	if language, ok := ingestFilenameLanguages[path.Base(p)]; ok {
		return language, true
	}
	language, ok := ingestExtensionLanguages[strings.ToLower(path.Ext(p))]
	return language, ok
}

// decodeTextBlob returns the blob as text, or false when it is not valid
// UTF-8 text (binary files, or content PostgreSQL TEXT cannot store).
func decodeTextBlob(blob *github.RepositoryBlob) (string, bool) {
	if blob == nil {
		return "", false
	}

	raw := blob.Content
	if blob.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(
			strings.NewReplacer("\n", "", "\r", "").Replace(raw),
		)
		if err != nil {
			return "", false
		}
		raw = string(decoded)
	}

	if raw == "" || !utf8.ValidString(raw) || strings.ContainsRune(raw, 0) {
		return "", false
	}

	return raw, true
}

var _ RepositoryIngestionService = (*RepositoryIngestionServiceImpl)(nil)
