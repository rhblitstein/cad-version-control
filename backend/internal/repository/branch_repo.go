package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/rhblitstein/cad-version-control/internal/models"
)

type BranchRepository struct {
	db    *sql.DB
	cache *RedisClient
}

func NewBranchRepository(db *sql.DB, cache *RedisClient) *BranchRepository {
	return &BranchRepository{
		db:    db,
		cache: cache,
	}
}

func (r *BranchRepository) Create(ctx context.Context, branch *models.Branch) error {
	query := `
		INSERT INTO branches (id, project_id, name, head_commit_id, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at
	`

	branch.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		branch.ID,
		branch.ProjectID,
		branch.Name,
		branch.HeadCommitID,
	).Scan(&branch.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

func (r *BranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Branch, error) {
	query := `
		SELECT id, project_id, name, head_commit_id, created_at
		FROM branches
		WHERE id = $1
	`

	var branch models.Branch
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&branch.ID,
		&branch.ProjectID,
		&branch.Name,
		&branch.HeadCommitID,
		&branch.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("branch not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}

	return &branch, nil
}

func (r *BranchRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.Branch, error) {
	query := `
		SELECT id, project_id, name, head_commit_id, created_at
		FROM branches
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	defer rows.Close()

	var branches []models.Branch
	for rows.Next() {
		var b models.Branch
		err := rows.Scan(
			&b.ID,
			&b.ProjectID,
			&b.Name,
			&b.HeadCommitID,
			&b.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan branch: %w", err)
		}
		branches = append(branches, b)
	}

	return branches, nil
}

func (r *BranchRepository) UpdateHead(ctx context.Context, branchID, commitID uuid.UUID) error {
	query := `
		UPDATE branches
		SET head_commit_id = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, commitID, branchID)
	if err != nil {
		return fmt.Errorf("failed to update branch head: %w", err)
	}

	cacheKey := fmt.Sprintf("branch:%s", branchID)
	r.cache.Del(ctx, cacheKey)

	return nil
}
