package github

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return f(request)
}

func testGitHubAppConfig(t *testing.T) config.GitHubAppConfig {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return config.GitHubAppConfig{
		AppID:      "123456",
		PrivateKey: string(privateKeyPEM),
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d test status", status),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

type fakeClient struct{}

func (f *fakeClient) GetInstallation(
	ctx context.Context,
	installationID string,
) (*Installation, error) {
	return &Installation{ID: installationID}, nil
}

func (f *fakeClient) ListRepositories(
	ctx context.Context,
	installationID string,
) ([]Repository, error) {
	return nil, nil
}

func (f *fakeClient) GetLatestCommitSHA(
	ctx context.Context,
	installationID string,
	owner string,
	repository string,
	branch string,
) (string, error) {
	return "abc123", nil
}
func (f *fakeClient) GetTree(
	ctx context.Context,
	installationID string,
	owner string,
	repository string,
	treeSHA string,
) (*RepositoryTree, error) {
	return nil, nil
}

func (f *fakeClient) GetBlob(
	ctx context.Context,
	installationID string,
	owner string,
	repository string,
	blobSHA string,
) (*RepositoryBlob, error) {
	return nil, nil
}

func TestClientContract(t *testing.T) {
	var client Client = &fakeClient{}

	installation, err := client.GetInstallation(
		context.Background(),
		"installation-123",
	)
	if err != nil {
		t.Fatalf("get installation: %v", err)
	}

	if installation.ID != "installation-123" {
		t.Fatalf("expected installation ID, got %s", installation.ID)
	}
}

func TestClientGetInstallation(t *testing.T) {
	var gotRequest *http.Request
	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			gotRequest = request
			return jsonResponse(
				http.StatusOK,
				`{"id":12345,"account":{"login":"devflow-org"}}`,
			), nil
		}),
	}
	client, err := NewClient(
		testGitHubAppConfig(t),
		httpClient,
		"https://github.test/api",
	)
	if err != nil {
		t.Fatalf("create github client: %v", err)
	}

	installation, err := client.GetInstallation(
		context.Background(),
		"12345",
	)
	if err != nil {
		t.Fatalf("get installation: %v", err)
	}

	if installation.ID != "12345" || installation.AccountLogin != "devflow-org" {
		t.Fatalf("unexpected installation: %+v", installation)
	}

	if gotRequest.Method != http.MethodGet {
		t.Fatalf("expected GET, got %s", gotRequest.Method)
	}

	if gotRequest.URL.String() != "https://github.test/api/app/installations/12345" {
		t.Fatalf("unexpected URL: %s", gotRequest.URL)
	}

	if !strings.HasPrefix(gotRequest.Header.Get("Authorization"), "Bearer ") {
		t.Fatal("expected bearer authorization header")
	}

	if gotRequest.Header.Get("X-GitHub-Api-Version") != githubAPIVersion {
		t.Fatal("expected GitHub API version header")
	}
}

func TestClientListRepositoriesUsesInstallationToken(t *testing.T) {
	requests := make([]*http.Request, 0, 2)
	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			requests = append(requests, request)
			if request.URL.Path == "/app/installations/12345/access_tokens" {
				return jsonResponse(http.StatusCreated, `{"token":"installation-token"}`), nil
			}

			return jsonResponse(
				http.StatusOK,
				`{"repositories":[{"id":99,"name":"devflow","full_name":"devflow/api","default_branch":"main","html_url":"https://github.com/devflow/api","clone_url":"https://github.com/devflow/api.git","private":true,"owner":{"login":"devflow"}}]}`,
			), nil
		}),
	}
	client, err := NewClient(
		testGitHubAppConfig(t),
		httpClient,
		"https://github.test",
	)
	if err != nil {
		t.Fatalf("create github client: %v", err)
	}

	repositories, err := client.ListRepositories(
		context.Background(),
		"12345",
	)
	if err != nil {
		t.Fatalf("list repositories: %v", err)
	}

	if len(repositories) != 1 {
		t.Fatalf("expected one repository, got %d", len(repositories))
	}

	if repositories[0].ExternalID != "99" ||
		repositories[0].Owner != "devflow" ||
		!repositories[0].IsPrivate {
		t.Fatalf("unexpected repository: %+v", repositories[0])
	}

	if len(requests) != 2 {
		t.Fatalf("expected two requests, got %d", len(requests))
	}

	if requests[1].Header.Get("Authorization") != "Bearer installation-token" {
		t.Fatalf("expected installation token, got %s", requests[1].Header.Get("Authorization"))
	}
}
func TestClientGetTree(t *testing.T) {
	requests := make([]*http.Request, 0, 2)

	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			requests = append(requests, request)

			if request.URL.Path == "/app/installations/12345/access_tokens" {
				return jsonResponse(
					http.StatusCreated,
					`{"token":"installation-token"}`,
				), nil
			}

			return jsonResponse(
				http.StatusOK,
				`{
					"sha":"tree-sha-123",
					"truncated":false,
					"tree":[
						{
							"path":"README.md",
							"mode":"100644",
							"type":"blob",
							"sha":"blob-sha-1",
							"size":120
						},
						{
							"path":"internal",
							"mode":"040000",
							"type":"tree",
							"sha":"tree-sha-2"
						}
					]
				}`,
			), nil
		}),
	}

	client, err := NewClient(
		testGitHubAppConfig(t),
		httpClient,
		"https://github.test",
	)
	if err != nil {
		t.Fatalf("create github client: %v", err)
	}

	tree, err := client.GetTree(
		context.Background(),
		"12345",
		"devflow",
		"api",
		"commit-sha-123",
	)
	if err != nil {
		t.Fatalf("get tree: %v", err)
	}

	if tree.SHA != "tree-sha-123" {
		t.Fatalf("expected tree SHA tree-sha-123, got %s", tree.SHA)
	}

	if tree.Truncated {
		t.Fatal("expected tree to not be truncated")
	}

	if len(tree.Entries) != 2 {
		t.Fatalf("expected two tree entries, got %d", len(tree.Entries))
	}

	if tree.Entries[0].Path != "README.md" ||
		tree.Entries[0].Type != "blob" ||
		tree.Entries[0].SHA != "blob-sha-1" ||
		tree.Entries[0].Size != 120 {
		t.Fatalf("unexpected first tree entry: %+v", tree.Entries[0])
	}

	if len(requests) != 2 {
		t.Fatalf("expected two requests, got %d", len(requests))
	}

	if requests[1].URL.String() !=
		"https://github.test/repos/devflow/api/git/trees/commit-sha-123?recursive=1" {
		t.Fatalf("unexpected tree URL: %s", requests[1].URL)
	}

	if requests[1].Header.Get("Authorization") != "Bearer installation-token" {
		t.Fatalf("expected installation token, got %s",
			requests[1].Header.Get("Authorization"))
	}
}

func TestClientGetBlob(t *testing.T) {
	requests := make([]*http.Request, 0, 2)

	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			requests = append(requests, request)

			if request.URL.Path == "/app/installations/12345/access_tokens" {
				return jsonResponse(
					http.StatusCreated,
					`{"token":"installation-token"}`,
				), nil
			}

			return jsonResponse(
				http.StatusOK,
				`{
					"sha":"blob-sha-123",
					"node_id":"abc",
					"size":42,
					"url":"https://api.github.com/repos/devflow/api/git/blobs/blob-sha-123",
					"content":"package main\n\nfunc main() {}\n",
					"encoding":"utf-8"
				}`,
			), nil
		}),
	}

	client, err := NewClient(
		testGitHubAppConfig(t),
		httpClient,
		"https://github.test",
	)
	if err != nil {
		t.Fatalf("create github client: %v", err)
	}

	blob, err := client.GetBlob(
		context.Background(),
		"12345",
		"devflow",
		"api",
		"blob-sha-123",
	)
	if err != nil {
		t.Fatalf("get blob: %v", err)
	}

	if blob.SHA != "blob-sha-123" {
		t.Fatalf("expected blob SHA blob-sha-123, got %s", blob.SHA)
	}

	if blob.Encoding != "utf-8" {
		t.Fatalf("expected utf-8 encoding, got %s", blob.Encoding)
	}

	if blob.Size != 42 {
		t.Fatalf("expected size 42, got %d", blob.Size)
	}

	if blob.Content != "package main\n\nfunc main() {}\n" {
		t.Fatalf("unexpected blob content: %q", blob.Content)
	}

	if len(requests) != 2 {
		t.Fatalf("expected two requests, got %d", len(requests))
	}

	if requests[1].URL.String() !=
		"https://github.test/repos/devflow/api/git/blobs/blob-sha-123" {
		t.Fatalf("unexpected blob URL: %s", requests[1].URL)
	}

	if requests[1].Header.Get("Authorization") != "Bearer installation-token" {
		t.Fatalf("expected installation token, got %s",
			requests[1].Header.Get("Authorization"))
	}
}

func TestClientUnexpectedStatus(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusUnauthorized, `{"message":"bad credentials"}`), nil
		}),
	}
	client, err := NewClient(
		testGitHubAppConfig(t),
		httpClient,
		"https://github.test",
	)
	if err != nil {
		t.Fatalf("create github client: %v", err)
	}

	_, err = client.GetInstallation(context.Background(), "12345")
	if !errors.Is(err, ErrUnexpectedStatus) {
		t.Fatalf("expected unexpected status error, got %v", err)
	}
}
