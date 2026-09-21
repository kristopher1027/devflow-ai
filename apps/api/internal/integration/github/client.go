package github

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/config"
)

const (
	defaultBaseURL   = "https://api.github.com"
	githubAPIVersion = "2022-11-28"
)

var (
	ErrUnexpectedStatus  = errors.New("github API returned unexpected status")
	ErrPrivateKeyInvalid = errors.New("github app private key is invalid")
)

type Installation struct {
	ID           string
	AccountLogin string
}

type Repository struct {
	ExternalID    string
	Owner         string
	Name          string
	FullName      string
	DefaultBranch string
	HTMLURL       string
	CloneURL      string
	IsPrivate     bool
}

type Client interface {
	GetInstallation(
		ctx context.Context,
		installationID string,
	) (*Installation, error)

	ListRepositories(
		ctx context.Context,
		installationID string,
	) ([]Repository, error)
}

type unavailableClient struct {
	err error
}

func NewUnavailableClient(err error) Client {
	return &unavailableClient{err: err}
}

func (c *unavailableClient) GetInstallation(
	ctx context.Context,
	installationID string,
) (*Installation, error) {
	return nil, c.err
}

func (c *unavailableClient) ListRepositories(
	ctx context.Context,
	installationID string,
) ([]Repository, error) {
	return nil, c.err
}

func (c *appClient) GetLatestCommitSHA(
	ctx context.Context,
	installationID string,
	owner string,
	repository string,
	branch string,
) (string, error) {

	token, err := c.createInstallationToken(
		ctx,
		installationID,
	)
	if err != nil {
		return "", err
	}

	request, err := c.newRequest(
		ctx,
		http.MethodGet,
		"/repos/"+owner+"/"+repository+"/commits/"+branch,
		"",
		token,
	)
	if err != nil {
		return "", err
	}

	var response struct {
		SHA string `json:"sha"`
	}

	if err := c.doJSON(request, &response); err != nil {
		return "", err
	}

	if response.SHA == "" {
		return "", errors.New("github commit SHA missing")
	}

	return response.SHA, nil
}

type appClient struct {
	appID      string
	privateKey *rsa.PrivateKey
	httpClient *http.Client
	baseURL    string
}

func NewClient(
	appConfig config.GitHubAppConfig,
	httpClient *http.Client,
	baseURL string,
) (Client, error) {
	if err := appConfig.Validate(); err != nil {
		return nil, err
	}

	privateKey, err := parsePrivateKey(appConfig.PrivateKey)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	return &appClient{
		appID:      appConfig.AppID,
		privateKey: privateKey,
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}, nil
}

func (c *appClient) GetInstallation(
	ctx context.Context,
	installationID string,
) (*Installation, error) {
	request, err := c.newRequest(
		ctx,
		http.MethodGet,
		"/app/installations/"+installationID,
		"",
		c.appJWT(),
	)
	if err != nil {
		return nil, err
	}

	var response struct {
		ID      int64 `json:"id"`
		Account struct {
			Login string `json:"login"`
		} `json:"account"`
	}
	if err := c.doJSON(request, &response); err != nil {
		return nil, err
	}

	return &Installation{
		ID:           fmt.Sprintf("%d", response.ID),
		AccountLogin: response.Account.Login,
	}, nil
}

func (c *appClient) ListRepositories(
	ctx context.Context,
	installationID string,
) ([]Repository, error) {
	token, err := c.createInstallationToken(ctx, installationID)
	if err != nil {
		return nil, err
	}

	request, err := c.newRequest(
		ctx,
		http.MethodGet,
		"/installation/repositories",
		"",
		token,
	)
	if err != nil {
		return nil, err
	}

	var response struct {
		Repositories []struct {
			ID            int64  `json:"id"`
			Name          string `json:"name"`
			FullName      string `json:"full_name"`
			DefaultBranch string `json:"default_branch"`
			HTMLURL       string `json:"html_url"`
			CloneURL      string `json:"clone_url"`
			Private       bool   `json:"private"`
			Owner         struct {
				Login string `json:"login"`
			} `json:"owner"`
		} `json:"repositories"`
	}
	if err := c.doJSON(request, &response); err != nil {
		return nil, err
	}

	repositories := make([]Repository, 0, len(response.Repositories))
	for _, repository := range response.Repositories {
		repositories = append(repositories, Repository{
			ExternalID:    fmt.Sprintf("%d", repository.ID),
			Owner:         repository.Owner.Login,
			Name:          repository.Name,
			FullName:      repository.FullName,
			DefaultBranch: repository.DefaultBranch,
			HTMLURL:       repository.HTMLURL,
			CloneURL:      repository.CloneURL,
			IsPrivate:     repository.Private,
		})
	}

	return repositories, nil
}

func (c *appClient) createInstallationToken(
	ctx context.Context,
	installationID string,
) (string, error) {
	request, err := c.newRequest(
		ctx,
		http.MethodPost,
		"/app/installations/"+installationID+"/access_tokens",
		"",
		c.appJWT(),
	)
	if err != nil {
		return "", err
	}

	var response struct {
		Token string `json:"token"`
	}
	if err := c.doJSON(request, &response); err != nil {
		return "", err
	}

	if response.Token == "" {
		return "", errors.New("github installation token is missing")
	}

	return response.Token, nil
}

func (c *appClient) newRequest(
	ctx context.Context,
	method string,
	path string,
	body string,
	token string,
) (*http.Request, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		method,
		c.baseURL+path,
		strings.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", githubAPIVersion)
	request.Header.Set("User-Agent", "devflow-ai")
	request.Header.Set("Authorization", "Bearer "+token)

	return request, nil
}

func (c *appClient) doJSON(
	request *http.Request,
	value any,
) error {
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("request github API: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: %s", ErrUnexpectedStatus, response.Status)
	}

	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(value); err != nil {
		return fmt.Errorf("decode github API response: %w", err)
	}

	return nil
}

func (c *appClient) appJWT() string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	now := time.Now().Unix()
	claims := fmt.Sprintf(`{"iat":%d,"exp":%d,"iss":"%s"}`, now-60, now+540, c.appID)
	encodedClaims := base64.RawURLEncoding.EncodeToString([]byte(claims))
	unsigned := header + "." + encodedClaims

	hash := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return ""
	}

	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func parsePrivateKey(value string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, ErrPrivateKeyInvalid
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPrivateKeyInvalid, err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrPrivateKeyInvalid
	}

	return rsaKey, nil
}
