package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/kristopher1027/devflow-ai/internal/database"
	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var ErrWorkspaceMemberNotFound = errors.New(
	"workspace member not found",
)

type WorkspaceMemberRepository interface {
	Create(
		ctx context.Context,
		member *domain.WorkspaceMember,
	) error

	Find(
		ctx context.Context,
		workspaceID string,
		userID string,
	) (*domain.WorkspaceMember, error)

	ListByWorkspaceID(
		ctx context.Context,
		workspaceID string,
	) ([]*domain.WorkspaceMember, error)

	UpdateRole(
		ctx context.Context,
		workspaceID string,
		userID string,
		role string,
	) error

	Delete(
		ctx context.Context,
		workspaceID string,
		userID string,
	) error
}


type PostgresWorkspaceMemberRepository struct {
	db *database.DB
}


func NewWorkspaceMemberRepository(
	db *database.DB,
) WorkspaceMemberRepository {

	return &PostgresWorkspaceMemberRepository{
		db: db,
	}
}


func (r *PostgresWorkspaceMemberRepository) Create(
	ctx context.Context,
	member *domain.WorkspaceMember,
) error {

	query := `
		INSERT INTO workspace_members (
			workspace_id,
			user_id,
			role,
			created_at
		)
		VALUES ($1,$2,$3,$4)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		member.WorkspaceID,
		member.UserID,
		member.Role,
		member.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"create workspace member: %w",
			err,
		)
	}

	return nil
}


func (r *PostgresWorkspaceMemberRepository) Find(
	ctx context.Context,
	workspaceID string,
	userID string,
) (*domain.WorkspaceMember, error) {

	query := `
		SELECT
			workspace_id,
			user_id,
			role,
			created_at
		FROM workspace_members
		WHERE workspace_id=$1
		AND user_id=$2
	`

	var member domain.WorkspaceMember


	err := r.db.Pool.QueryRow(
		ctx,
		query,
		workspaceID,
		userID,
	).Scan(
		&member.WorkspaceID,
		&member.UserID,
		&member.Role,
		&member.CreatedAt,
	)


	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkspaceMemberNotFound
		}

		return nil, fmt.Errorf(
			"find workspace member: %w",
			err,
		)
	}


	return &member,nil
}



func (r *PostgresWorkspaceMemberRepository) ListByWorkspaceID(
	ctx context.Context,
	workspaceID string,
) ([]*domain.WorkspaceMember,error) {


	query := `
		SELECT
			workspace_id,
			user_id,
			role,
			created_at
		FROM workspace_members
		WHERE workspace_id=$1
		ORDER BY created_at ASC
	`


	rows,err := r.db.Pool.Query(
		ctx,
		query,
		workspaceID,
	)


	if err != nil {
		return nil,fmt.Errorf(
			"list workspace members: %w",
			err,
		)
	}


	defer rows.Close()


	members := make(
		[]*domain.WorkspaceMember,
		0,
	)


	for rows.Next(){

		var member domain.WorkspaceMember


		err := rows.Scan(
			&member.WorkspaceID,
			&member.UserID,
			&member.Role,
			&member.CreatedAt,
		)


		if err != nil {
			return nil,fmt.Errorf(
				"scan workspace member: %w",
				err,
			)
		}


		members = append(
			members,
			&member,
		)
	}



	if err := rows.Err(); err != nil {

		return nil,fmt.Errorf(
			"iterate workspace members: %w",
			err,
		)
	}


	return members,nil
}




func (r *PostgresWorkspaceMemberRepository) UpdateRole(
	ctx context.Context,
	workspaceID string,
	userID string,
	role string,
) error {


	query := `
		UPDATE workspace_members
		SET role=$1
		WHERE workspace_id=$2
		AND user_id=$3
	`


	result,err := r.db.Pool.Exec(
		ctx,
		query,
		role,
		workspaceID,
		userID,
	)


	if err != nil {

		return fmt.Errorf(
			"update workspace member role: %w",
			err,
		)
	}



	if result.RowsAffected()==0 {

		return ErrWorkspaceMemberNotFound
	}


	return nil
}





func (r *PostgresWorkspaceMemberRepository) Delete(
	ctx context.Context,
	workspaceID string,
	userID string,
) error {


	query := `
		DELETE FROM workspace_members
		WHERE workspace_id=$1
		AND user_id=$2
	`


	result,err := r.db.Pool.Exec(
		ctx,
		query,
		workspaceID,
		userID,
	)


	if err != nil {

		return fmt.Errorf(
			"delete workspace member: %w",
			err,
		)
	}



	if result.RowsAffected()==0 {

		return ErrWorkspaceMemberNotFound
	}


	return nil
}