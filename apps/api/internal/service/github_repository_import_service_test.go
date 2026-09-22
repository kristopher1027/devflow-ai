package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	githubintegration "github.com/kristopher1027/devflow-ai/internal/integration/github"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type fakeGitHubClient struct {
	repositories      []githubintegration.Repository
	err               error
	gotInstallationID string
}

func (f *fakeGitHubClient) GetInstallation(
	ctx context.Context,
	installationID string,
) (*githubintegration.Installation, error) {
	return nil, nil
}

func (f *fakeGitHubClient) ListRepositories(
	ctx context.Context,
	installationID string,
) ([]githubintegration.Repository, error) {
	f.gotInstallationID = installationID
	if f.err != nil {
		return nil, f.err
	}
	return f.repositories, nil
}
func (f *fakeGitHubClient) GetLatestCommitSHA(
	ctx context.Context,
	installationID string,
	owner string,
	repository string,
	branch string,
) (string, error) {
	return "", nil
}

type fakeGitHubConnectionServiceForImport struct {
	connection     *domain.GitHubConnection
	err            error
	gotRequesterID string
	gotWorkspaceID string
}

func (f *fakeGitHubConnectionServiceForImport) Create(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	installationID string,
	accountLogin string,
) (*domain.GitHubConnection, error) {
	return nil, nil
}

func (f *fakeGitHubConnectionServiceForImport) FindByWorkspaceID(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) (*domain.GitHubConnection, error) {
	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID
	if f.err != nil {
		return nil, f.err
	}
	return f.connection, nil
}

func (f *fakeGitHubConnectionServiceForImport) UpdateStatus(
	ctx context.Context,
	requesterID string,
	id string,
	status string,
) error {
	return nil
}

type fakeRepositoryServiceForImport struct {
	repositories []*domain.Repository
	err          error
	createCalls  []githubRepositoryCreateCall
}

type githubRepositoryCreateCall struct {
	requesterID   string
	projectID     string
	provider      string
	externalID    string
	owner         string
	name          string
	fullName      string
	defaultBranch string
	htmlURL       string
	cloneURL      string
	isPrivate     bool
}

func (f *fakeRepositoryServiceForImport) Create(
	ctx context.Context,
	requesterID string,
	projectID string,
	provider string,
	externalID string,
	owner string,
	name string,
	fullName string,
	defaultBranch string,
	htmlURL string,
	cloneURL string,
	isPrivate bool,
) (*domain.Repository, error) {
	f.createCalls = append(f.createCalls, githubRepositoryCreateCall{
		requesterID:   requesterID,
		projectID:     projectID,
		provider:      provider,
		externalID:    externalID,
		owner:         owner,
		name:          name,
		fullName:      fullName,
		defaultBranch: defaultBranch,
		htmlURL:       htmlURL,
		cloneURL:      cloneURL,
		isPrivate:     isPrivate,
	})
	if f.err != nil {
		return nil, f.err
	}
	created := &domain.Repository{ExternalID: externalID, ProjectID: projectID}
	f.repositories = append(f.repositories, created)
	return created, nil
}

func (f *fakeRepositoryServiceForImport) FindByID(context.Context, string, string) (*domain.Repository, error) {
	return nil, nil
}

func (f *fakeRepositoryServiceForImport) ListByProjectID(context.Context, string, string) ([]*domain.Repository, error) {
	return nil, nil
}

func (f *fakeRepositoryServiceForImport) Delete(context.Context, string, string) error {
	return nil
}

func TestGitHubRepositoryImportServiceImportRepositories(t *testing.T) {
	githubClient := &fakeGitHubClient{
		repositories: []githubintegration.Repository{
			{
				ExternalID:    "99",
				Owner:         "devflow",
				Name:          "api",
				FullName:      "devflow/api",
				DefaultBranch: "main",
				HTMLURL:       "https://github.com/devflow/api",
				CloneURL:      "https://github.com/devflow/api.git",
				IsPrivate:     true,
			},
		},
	}
	connectionService := &fakeGitHubConnectionServiceForImport{
		connection: &domain.GitHubConnection{InstallationID: "installation-123"},
	}
	repositoryService := &fakeRepositoryServiceForImport{}
	service := NewGitHubRepositoryImportService(
		&fakeProjectRepository{project: testProject()},
		connectionService,
		githubClient,
		repositoryService,
	)

	imported, err := service.ImportRepositories(
		context.Background(),
		"user-123",
		"project-123",
	)
	if err != nil {
		t.Fatalf("import repositories: %v", err)
	}

	if len(imported) != 1 || len(repositoryService.createCalls) != 1 {
		t.Fatalf("expected one imported repository")
	}

	call := repositoryService.createCalls[0]
	if call.provider != "github" || call.projectID != "project-123" || call.externalID != "99" {
		t.Fatalf("unexpected create call: %+v", call)
	}
	if call.owner != "devflow" || call.fullName != "devflow/api" || !call.isPrivate {
		t.Fatalf("unexpected repository metadata: %+v", call)
	}
	if githubClient.gotInstallationID != "installation-123" {
		t.Fatalf("expected installation ID, got %s", githubClient.gotInstallationID)
	}
}

func TestGitHubRepositoryImportServiceSkipsDuplicates(t *testing.T) {
	repositoryService := &fakeRepositoryServiceForImport{
		err: ErrRepositoryAlreadyExists,
	}
	service := NewGitHubRepositoryImportService(
		&fakeProjectRepository{project: testProject()},
		&fakeGitHubConnectionServiceForImport{
			connection: &domain.GitHubConnection{InstallationID: "installation-123"},
		},
		&fakeGitHubClient{repositories: []githubintegration.Repository{{ExternalID: "99"}}},
		repositoryService,
	)

	imported, err := service.ImportRepositories(context.Background(), "user-123", "project-123")
	if err != nil {
		t.Fatalf("import duplicate repository: %v", err)
	}
	if len(imported) != 0 {
		t.Fatalf("expected duplicate to be skipped")
	}
}

func TestGitHubRepositoryImportServicePropagatesClientError(t *testing.T) {
	clientErr := errors.New("github unavailable")
	service := NewGitHubRepositoryImportService(
		&fakeProjectRepository{project: testProject()},
		&fakeGitHubConnectionServiceForImport{
			connection: &domain.GitHubConnection{InstallationID: "installation-123"},
		},
		&fakeGitHubClient{err: clientErr},
		&fakeRepositoryServiceForImport{},
	)

	_, err := service.ImportRepositories(context.Background(), "user-123", "project-123")
	if !errors.Is(err, clientErr) {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestGitHubRepositoryImportServicePropagatesConnectionError(t *testing.T) {
	connectionErr := errors.New("connection unavailable")
	service := NewGitHubRepositoryImportService(
		&fakeProjectRepository{project: testProject()},
		&fakeGitHubConnectionServiceForImport{err: connectionErr},
		&fakeGitHubClient{},
		&fakeRepositoryServiceForImport{},
	)

	_, err := service.ImportRepositories(context.Background(), "user-123", "project-123")
	if !errors.Is(err, connectionErr) {
		t.Fatalf("expected connection error, got %v", err)
	}
}

var _ GitHubRepositoryImportService = (*GitHubRepositoryImportServiceImpl)(nil)
var _ repository.GitHubConnectionRepository
