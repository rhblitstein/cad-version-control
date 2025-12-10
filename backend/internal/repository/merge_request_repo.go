package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/rhblitstein/cad-version-control/internal/models"
)

type MergeRequestRepository struct {
	db *sql.DB
}

func NewMergeRequestRepository(db *sql.DB) *MergeRequestRepository {
	return &MergeRequestRepository{db: db}
}

func (r *MergeRequestRepository) Create(ctx context.Context, mr *models.MergeRequest) error {
	query := `
		INSERT INTO merge_requests (id, project_id, source_branch_id, target_branch_id, title, description, status, author, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	mr.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		mr.ID,
		mr.ProjectID,
		mr.SourceBranchID,
		mr.TargetBranchID,
		mr.Title,
		mr.Description,
		mr.Status,
		mr.Author,
	).Scan(&mr.CreatedAt, &mr.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create merge request: %w", err)
	}

	return nil
}

func (r *MergeRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.MergeRequest, error) {
	query := `
		SELECT id, project_id, source_branch_id, target_branch_id, title, description, status, author, created_at, updated_at, merged_at
		FROM merge_requests
		WHERE id = $1
	`

	var mr models.MergeRequest
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&mr.ID,
		&mr.ProjectID,
		&mr.SourceBranchID,
		&mr.TargetBranchID,
		&mr.Title,
		&mr.Description,
		&mr.Status,
		&mr.Author,
		&mr.CreatedAt,
		&mr.UpdatedAt,
		&mr.MergedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("merge request not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get merge request: %w", err)
	}

	return &mr, nil
}

func (r *MergeRequestRepository) List(ctx context.Context, projectID *uuid.UUID, status string) ([]models.MergeRequest, error) {
	query := `
		SELECT id, project_id, source_branch_id, target_branch_id, title, description, status, author, created_at, updated_at, merged_at
		FROM merge_requests
		WHERE ($1::uuid IS NULL OR project_id = $1)
		  AND ($2 = '' OR status = $2)
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list merge requests: %w", err)
	}
	defer rows.Close()

	var mrs []models.MergeRequest
	for rows.Next() {
		var mr models.MergeRequest
		err := rows.Scan(
			&mr.ID,
			&mr.ProjectID,
			&mr.SourceBranchID,
			&mr.TargetBranchID,
			&mr.Title,
			&mr.Description,
			&mr.Status,
			&mr.Author,
			&mr.CreatedAt,
			&mr.UpdatedAt,
			&mr.MergedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merge request: %w", err)
		}
		mrs = append(mrs, mr)
	}

	return mrs, nil
}

func (r *MergeRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE merge_requests
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update merge request status: %w", err)
	}

	return nil
}

// Conflicts
func (r *MergeRequestRepository) CreateConflict(ctx context.Context, conflict *models.MergeConflict) error {
	query := `
		INSERT INTO merge_conflicts (id, merge_request_id, file_id, source_version_id, target_version_id, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	conflict.ID = uuid.New()

	_, err := r.db.ExecContext(ctx, query,
		conflict.ID,
		conflict.MergeRequestID,
		conflict.FileID,
		conflict.SourceVersionID,
		conflict.TargetVersionID,
		conflict.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to create conflict: %w", err)
	}

	return nil
}

func (r *MergeRequestRepository) GetConflicts(ctx context.Context, mrID uuid.UUID) ([]models.MergeConflict, error) {
	query := `
		SELECT id, merge_request_id, file_id, source_version_id, target_version_id, status, 
	       	COALESCE(resolution_notes, '') as resolution_notes, resolved_at
		FROM merge_conflicts
		WHERE merge_request_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, mrID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflicts: %w", err)
	}
	defer rows.Close()

	var conflicts []models.MergeConflict
	for rows.Next() {
		var c models.MergeConflict
		err := rows.Scan(
			&c.ID,
			&c.MergeRequestID,
			&c.FileID,
			&c.SourceVersionID,
			&c.TargetVersionID,
			&c.Status,
			&c.ResolutionNotes,
			&c.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conflict: %w", err)
		}
		conflicts = append(conflicts, c)
	}

	return conflicts, nil
}

func (r *MergeRequestRepository) GetConflictByID(ctx context.Context, id uuid.UUID) (*models.MergeConflict, error) {
	query := `
		SELECT id, merge_request_id, file_id, source_version_id, target_version_id, status, 
		       COALESCE(resolution_notes, '') as resolution_notes, resolved_at
		FROM merge_conflicts
		WHERE id = $1
	`

	var c models.MergeConflict
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.MergeRequestID,
		&c.FileID,
		&c.SourceVersionID,
		&c.TargetVersionID,
		&c.Status,
		&c.ResolutionNotes, // Now handles empty string
		&c.ResolvedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conflict not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conflict: %w", err)
	}

	return &c, nil
}

func (r *MergeRequestRepository) ResolveConflict(ctx context.Context, id uuid.UUID, notes string) error {
	query := `
		UPDATE merge_conflicts
		SET status = 'resolved', resolution_notes = $1, resolved_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, notes, id)
	if err != nil {
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	return nil
}

// Comments
func (r *MergeRequestRepository) CreateComment(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (id, merge_request_id, author, content, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at
	`

	comment.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		comment.ID,
		comment.MergeRequestID,
		comment.Author,
		comment.Content,
	).Scan(&comment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

func (r *MergeRequestRepository) GetComments(ctx context.Context, mrID uuid.UUID) ([]models.Comment, error) {
	query := `
		SELECT id, merge_request_id, author, content, created_at
		FROM comments
		WHERE merge_request_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, mrID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(
			&c.ID,
			&c.MergeRequestID,
			&c.Author,
			&c.Content,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	return comments, nil
}

// Approvals
func (r *MergeRequestRepository) CreateApproval(ctx context.Context, approval *models.Approval) error {
	query := `
		INSERT INTO approvals (id, merge_request_id, approver, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING created_at
	`

	approval.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		approval.ID,
		approval.MergeRequestID,
		approval.Approver,
	).Scan(&approval.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create approval: %w", err)
	}

	return nil
}

func (r *MergeRequestRepository) GetApprovals(ctx context.Context, mrID uuid.UUID) ([]models.Approval, error) {
	query := `
		SELECT id, merge_request_id, approver, created_at
		FROM approvals
		WHERE merge_request_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, mrID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approvals: %w", err)
	}
	defer rows.Close()

	var approvals []models.Approval
	for rows.Next() {
		var a models.Approval
		err := rows.Scan(
			&a.ID,
			&a.MergeRequestID,
			&a.Approver,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan approval: %w", err)
		}
		approvals = append(approvals, a)
	}

	return approvals, nil
}
