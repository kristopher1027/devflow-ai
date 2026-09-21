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
