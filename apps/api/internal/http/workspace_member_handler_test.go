package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type fakeWorkspaceMemberService struct {
	member  *domain.WorkspaceMember
	members []*domain.WorkspaceMember
	err     error

	gotRequesterID string
	gotWorkspaceID string
	gotUserID      string
	gotRole        string
}

func (f *fakeWorkspaceMemberService) UpdateRole(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
	role string,
) error {

	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID
	f.gotUserID = userID
	f.gotRole = role

	if f.err != nil {
		return f.err
	}

	if f.member != nil {
		f.member.Role = role
	}

	return nil
}
func (f *fakeWorkspaceMemberService) Add(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
	role string,
) (*domain.WorkspaceMember, error) {
	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID
	f.gotUserID = userID
	f.gotRole = role

	if f.err != nil {
		return nil, f.err
	}

	return f.member, nil
}

func (f *fakeWorkspaceMemberService) Find(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
) (*domain.WorkspaceMember, error) {
	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID
	f.gotUserID = userID

	if f.err != nil {
		return nil, f.err
	}

	return f.member, nil
}

func (f *fakeWorkspaceMemberService) ListByWorkspaceID(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) ([]*domain.WorkspaceMember, error) {
	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID

	if f.err != nil {
		return nil, f.err
	}

	return f.members, nil
}

func (f *fakeWorkspaceMemberService) Remove(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
) error {
	f.gotRequesterID = requesterID
	f.gotWorkspaceID = workspaceID
	f.gotUserID = userID

	return f.err
}

var _ service.WorkspaceMemberService = (*fakeWorkspaceMemberService)(nil)

func TestWorkspaceMemberHandlerAdd(t *testing.T) {
	now := time.Now()

	member := &domain.WorkspaceMember{
		WorkspaceID: "workspace-123",
		UserID:      "user-456",
		Role:        service.WorkspaceMemberRoleMember,
		CreatedAt:   now,
	}

	memberService := &fakeWorkspaceMemberService{
		member: member,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/members",
		`{"user_id":"user-456","role":"member"}`,
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if memberService.gotRequesterID != "user-owner" {
		t.Fatalf(
			"expected requester ID %q, got %q",
			"user-owner",
			memberService.gotRequesterID,
		)
	}

	if memberService.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID %q, got %q",
			"workspace-123",
			memberService.gotWorkspaceID,
		)
	}

	if memberService.gotUserID != "user-456" {
		t.Fatalf(
			"expected user ID %q, got %q",
			"user-456",
			memberService.gotUserID,
		)
	}

	if memberService.gotRole != service.WorkspaceMemberRoleMember {
		t.Fatalf(
			"expected role %q, got %q",
			service.WorkspaceMemberRoleMember,
			memberService.gotRole,
		)
	}
}

func TestWorkspaceMemberHandlerAddRequiresAuth(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{}

	handler := NewWorkspaceMemberHandler(memberService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/workspaces/workspace-123/members",
		strings.NewReader(
			`{"user_id":"user-456","role":"member"}`,
		),
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerAddInvalidBody(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/members",
		`invalid-json`,
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerAddMissingUserID(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/members",
		`{"user_id":"","role":"member"}`,
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerAddRoleRequired(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: service.ErrWorkspaceMemberRoleRequired,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/members",
		`{"user_id":"user-456","role":""}`,
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerAddInvalidRole(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: service.ErrInvalidWorkspaceMemberRole,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/members",
		`{"user_id":"user-456","role":"superadmin"}`,
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerAddAdminCannotAssignRole(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: service.ErrWorkspaceMemberAdminCannotAssignRole,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/members",
		`{"user_id":"user-456","role":"admin"}`,
		"user-admin",
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerAddUnauthorized(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: service.ErrWorkspaceMemberUnauthorized,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/members",
		`{"user_id":"user-456","role":"member"}`,
		"user-member",
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerAddWorkspaceNotFound(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: repository.ErrWorkspaceNotFound,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/missing/members",
		`{"user_id":"user-456","role":"member"}`,
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerAddInternalError(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: errors.New("database failure"),
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodPost,
		"/workspaces/workspace-123/members",
		`{"user_id":"user-456","role":"member"}`,
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Add(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerList(t *testing.T) {
	now := time.Now()

	members := []*domain.WorkspaceMember{
		{
			WorkspaceID: "workspace-123",
			UserID:      "user-owner",
			Role:        service.WorkspaceMemberRoleOwner,
			CreatedAt:   now,
		},
		{
			WorkspaceID: "workspace-123",
			UserID:      "user-member",
			Role:        service.WorkspaceMemberRoleMember,
			CreatedAt:   now,
		},
	}

	memberService := &fakeWorkspaceMemberService{
		members: members,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/members",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if memberService.gotRequesterID != "user-owner" {
		t.Fatalf(
			"expected requester ID %q, got %q",
			"user-owner",
			memberService.gotRequesterID,
		)
	}

	if memberService.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID %q, got %q",
			"workspace-123",
			memberService.gotWorkspaceID,
		)
	}
}

func TestWorkspaceMemberHandlerListRequiresAuth(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{}

	handler := NewWorkspaceMemberHandler(memberService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/workspace-123/members",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerListUnauthorized(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: service.ErrWorkspaceMemberUnauthorized,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/members",
		"",
		"user-member",
	)

	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerListWorkspaceNotFound(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: repository.ErrWorkspaceNotFound,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/missing/members",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerGet(t *testing.T) {
	now := time.Now()

	member := &domain.WorkspaceMember{
		WorkspaceID: "workspace-123",
		UserID:      "user-member",
		Role:        service.WorkspaceMemberRoleMember,
		CreatedAt:   now,
	}

	memberService := &fakeWorkspaceMemberService{
		member: member,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/members/user-member",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if memberService.gotRequesterID != "user-owner" {
		t.Fatalf(
			"expected requester ID %q, got %q",
			"user-owner",
			memberService.gotRequesterID,
		)
	}

	if memberService.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID %q, got %q",
			"workspace-123",
			memberService.gotWorkspaceID,
		)
	}

	if memberService.gotUserID != "user-member" {
		t.Fatalf(
			"expected user ID %q, got %q",
			"user-member",
			memberService.gotUserID,
		)
	}
}

func TestWorkspaceMemberHandlerGetRequiresAuth(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{}

	handler := NewWorkspaceMemberHandler(memberService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/workspaces/workspace-123/members/user-member",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerGetUnauthorized(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: service.ErrWorkspaceMemberUnauthorized,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/members/user-member",
		"",
		"user-member",
	)

	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerGetWorkspaceNotFound(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: repository.ErrWorkspaceNotFound,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/missing/members/user-member",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerGetMemberNotFound(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: repository.ErrWorkspaceMemberNotFound,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodGet,
		"/workspaces/workspace-123/members/missing-user",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerRemove(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/workspaces/workspace-123/members/user-member",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Remove(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
			rec.Code,
		)
	}

	if memberService.gotRequesterID != "user-owner" {
		t.Fatalf(
			"expected requester ID %q, got %q",
			"user-owner",
			memberService.gotRequesterID,
		)
	}

	if memberService.gotWorkspaceID != "workspace-123" {
		t.Fatalf(
			"expected workspace ID %q, got %q",
			"workspace-123",
			memberService.gotWorkspaceID,
		)
	}

	if memberService.gotUserID != "user-member" {
		t.Fatalf(
			"expected user ID %q, got %q",
			"user-member",
			memberService.gotUserID,
		)
	}
}

func TestWorkspaceMemberHandlerRemoveRequiresAuth(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{}

	handler := NewWorkspaceMemberHandler(memberService)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/workspaces/workspace-123/members/user-member",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Remove(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerRemoveUnauthorized(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: service.ErrWorkspaceMemberUnauthorized,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/workspaces/workspace-123/members/user-member",
		"",
		"user-member",
	)

	rec := httptest.NewRecorder()

	handler.Remove(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerRemoveOwner(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: service.ErrWorkspaceOwnerCannotBeRemoved,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/workspaces/workspace-123/members/user-owner",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Remove(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerRemoveWorkspaceNotFound(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: repository.ErrWorkspaceNotFound,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/workspaces/missing/members/user-member",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Remove(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerRemoveMemberNotFound(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: repository.ErrWorkspaceMemberNotFound,
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/workspaces/workspace-123/members/missing-user",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Remove(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestWorkspaceMemberHandlerRemoveInternalError(t *testing.T) {
	memberService := &fakeWorkspaceMemberService{
		err: errors.New("database failure"),
	}

	handler := NewWorkspaceMemberHandler(memberService)

	req := authenticatedRequest(
		http.MethodDelete,
		"/workspaces/workspace-123/members/user-member",
		"",
		"user-owner",
	)

	rec := httptest.NewRecorder()

	handler.Remove(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
