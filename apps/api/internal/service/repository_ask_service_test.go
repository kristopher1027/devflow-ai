package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type askSearchServiceFake struct {
	resultsByTerm map[string][]*domain.RepositoryChunkSearchResult
	err           error
	queries       []string
}

func (f *askSearchServiceFake) Search(
	ctx context.Context,
	snapshotID string,
	query string,
	limit int,
) ([]*domain.RepositoryChunkSearchResult, error) {
	f.queries = append(f.queries, query)
	return f.resultsByTerm[query], f.err
}

func askChunk(
	id string,
	path string,
	start int,
	end int,
	content string,
) *domain.RepositoryChunkSearchResult {
	return &domain.RepositoryChunkSearchResult{
		Chunk: &domain.RepositoryChunk{
			ID: id, FileID: path, ChunkIndex: 0,
			StartLine: start, EndLine: end, Content: content,
		},
		FilePath: path,
		Score:    1,
	}
}

type askFixture struct {
	service   *RepositoryAskServiceImpl
	members   *fakeWorkspaceMemberRepositoryForImportJob
	snapshots *explainSnapshotRepository
	search    *askSearchServiceFake
	generator *fakeTextGenerator
}

func newAskFixture() *askFixture {
	fixture := &askFixture{
		members: &fakeWorkspaceMemberRepositoryForImportJob{
			member: &domain.WorkspaceMember{Role: WorkspaceMemberRoleMember},
		},
		snapshots: &explainSnapshotRepository{
			snapshots: []*domain.RepositorySnapshot{
				{ID: "snap-new", CommitSHA: "commit-new"},
				{ID: "snap-old", CommitSHA: "commit-old"},
			},
		},
		search: &askSearchServiceFake{
			resultsByTerm: map[string][]*domain.RepositoryChunkSearchResult{
				"authentication": {
					askChunk("c1", "auth.go", 1, 20, "func Authenticate() {}"),
				},
				"login": {
					askChunk("c1", "auth.go", 1, 20, "func Authenticate() {}"),
					askChunk("c2", "login.go", 5, 30, "func Login() {}"),
				},
			},
		},
		generator: &fakeTextGenerator{result: "Sessions are checked in auth.go."},
	}

	fixture.service = NewRepositoryAskService(
		&fakeRepositorySyncRepository{repository: &domain.Repository{
			ID: "repo-1", ProjectID: "project-1",
		}},
		&ingestProjectRepository{project: &domain.Project{
			ID: "project-1", WorkspaceID: "ws-1",
		}},
		fixture.members,
		fixture.snapshots,
		fixture.search,
		NewDefaultRepositoryRetrievalContextBuilder(),
		NewRepositoryAIContextPackager(),
		fixture.generator,
	).(*RepositoryAskServiceImpl)

	return fixture
}

func TestRepositoryAskAnswersFromRetrievedSources(t *testing.T) {
	fixture := newAskFixture()
	question := "How does authentication work with login?"

	got, err := fixture.service.Ask(context.Background(), "user-1", "repo-1", question)

	require.NoError(t, err)
	require.Equal(t, "Sessions are checked in auth.go.", got.Answer)
	require.Equal(t, "snap-new", got.SnapshotID)
	require.Equal(t, "commit-new", got.CommitSHA)
	require.ElementsMatch(t, []string{"authentication", "login"}, fixture.search.queries)
	require.Len(t, got.Sources, 2)
	require.Equal(t, "ws-1", fixture.members.findArgs.workspaceID)
	require.Equal(t, "user-1", fixture.members.findArgs.userID)
	require.Equal(t, askSystemPrompt, fixture.generator.system)
	require.Contains(t, fixture.generator.prompt, "SOURCE 1")
	require.Contains(t, fixture.generator.prompt, "auth.go")
	require.Contains(t, fixture.generator.prompt, "login.go")
	require.Contains(t, fixture.generator.prompt, "QUESTION:\n"+question)
}

func TestRepositoryAskValidatesQuestion(t *testing.T) {
	fixture := newAskFixture()

	_, err := fixture.service.Ask(context.Background(), "user-1", "repo-1", "   ")
	require.ErrorIs(t, err, ErrRepositoryAskQuestionRequired)

	_, err = fixture.service.Ask(
		context.Background(), "user-1", "repo-1",
		strings.Repeat("a", maxAskQuestionChars+1),
	)
	require.ErrorIs(t, err, ErrRepositoryAskQuestionTooLong)

	require.Empty(t, fixture.search.queries)
	require.Zero(t, fixture.generator.calls)
}

func TestRepositoryAskRejectsNonMembersAndUnknownRoles(t *testing.T) {
	fixture := newAskFixture()
	fixture.members.findErr = repository.ErrWorkspaceMemberNotFound

	_, err := fixture.service.Ask(context.Background(), "user-1", "repo-1", "login flow")
	require.ErrorIs(t, err, ErrRepositoryAskUnauthorized)

	fixture = newAskFixture()
	fixture.members.member = &domain.WorkspaceMember{Role: "viewer"}

	_, err = fixture.service.Ask(context.Background(), "user-1", "repo-1", "login flow")
	require.ErrorIs(t, err, ErrRepositoryAskUnauthorized)

	require.Empty(t, fixture.search.queries)
	require.Zero(t, fixture.generator.calls)
}

func TestRepositoryAskRequiresSnapshot(t *testing.T) {
	fixture := newAskFixture()
	fixture.snapshots.snapshots = nil

	_, err := fixture.service.Ask(context.Background(), "user-1", "repo-1", "login flow")

	require.ErrorIs(t, err, ErrRepositoryAskNoSnapshot)
	require.Zero(t, fixture.generator.calls)
}

func TestRepositoryAskReportsNoMatchesWithoutCallingTheModel(t *testing.T) {
	fixture := newAskFixture()

	_, err := fixture.service.Ask(
		context.Background(), "user-1", "repo-1", "kubernetes operators",
	)
	require.ErrorIs(t, err, ErrRepositoryAskNoMatches)

	// A question made only of stop words has nothing to search for.
	_, err = fixture.service.Ask(context.Background(), "user-1", "repo-1", "how does it work?")
	require.ErrorIs(t, err, ErrRepositoryAskNoMatches)

	require.Zero(t, fixture.generator.calls)
}

func TestRepositoryAskPropagatesSearchAndGeneratorErrors(t *testing.T) {
	fixture := newAskFixture()
	fixture.search.err = errors.New("search down")

	_, err := fixture.service.Ask(context.Background(), "user-1", "repo-1", "login flow")
	require.EqualError(t, err, "search down")

	fixture = newAskFixture()
	fixture.generator.err = errors.New("model unavailable")

	_, err = fixture.service.Ask(context.Background(), "user-1", "repo-1", "login")
	require.EqualError(t, err, "model unavailable")
}

func TestAskSearchTermsDropsStopWordsAndPrefersSpecificWords(t *testing.T) {
	terms := askSearchTerms(
		"How does the authentication middleware validate session tokens for every request in the API?",
	)

	require.LessOrEqual(t, len(terms), maxAskSearchTerms)
	require.Equal(t, "authentication", terms[0])
	for _, stop := range []string{"how", "does", "the", "for", "in", "every"} {
		require.NotContains(t, terms, stop)
	}
	require.Contains(t, terms, "middleware")
	require.Contains(t, terms, "validate")
}

func TestAskSearchTermsCapsAndDeduplicates(t *testing.T) {
	terms := askSearchTerms("alpha1 beta22 gamma333 delta4444 omega55555 sigma666666 kappa7777777 alpha1")

	require.Len(t, terms, maxAskSearchTerms)
	require.Equal(t, "kappa7777777", terms[0])
	require.NotContains(t, terms[1:], "kappa7777777")
}

func TestAskRetrieveRanksChunksMatchingMoreTermsFirst(t *testing.T) {
	fixture := newAskFixture()
	fixture.search.resultsByTerm = map[string][]*domain.RepositoryChunkSearchResult{
		"alphabet": {
			askChunk("only-alpha", "a.go", 1, 5, "alpha"),
			askChunk("both", "z.go", 1, 5, "alpha beta"),
		},
		"beta": {
			askChunk("both", "z.go", 1, 5, "alpha beta"),
		},
	}

	results, err := fixture.service.retrieve(
		context.Background(), "snap-1", "alphabet beta",
	)

	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, "both", results[0].Chunk.ID)
	require.Equal(t, []string{"alphabet", "beta"}, results[0].MatchedTerms)
	require.Equal(t, "only-alpha", results[1].Chunk.ID)
}

func TestAskRetrieveCapsTheNumberOfChunks(t *testing.T) {
	fixture := newAskFixture()
	many := make([]*domain.RepositoryChunkSearchResult, 0, maxAskChunks+5)
	for i := 0; i < maxAskChunks+5; i++ {
		many = append(many, askChunk(
			fmt.Sprintf("c%02d", i), fmt.Sprintf("f%02d.go", i), 1, 5, "x",
		))
	}
	fixture.search.resultsByTerm = map[string][]*domain.RepositoryChunkSearchResult{
		"database": many,
	}

	results, err := fixture.service.retrieve(context.Background(), "snap-1", "database")

	require.NoError(t, err)
	require.Len(t, results, maxAskChunks)
}
