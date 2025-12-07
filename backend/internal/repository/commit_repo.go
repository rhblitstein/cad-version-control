package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/rhblitstein/cad-version-control/internal/models"
)

type CommitRepository struct {
	db *sql.DB
}

func NewCommitRepository(db *sql.DB) *CommitRepository {
	return &CommitRepository{db: db}
}

func (r *CommitRepository) Create(ctx context.Context, commit *models.Commit) error {
	query := `
		INSERT INTO commits (id, project_id, branch_id, parent_commit_id, author, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING created_at
	`

	commit.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		commit.ID,
		commit.ProjectID,
		commit.BranchID,
		commit.ParentCommitID,
		commit.Author,
		commit.Message,
	).Scan(&commit.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

func (r *CommitRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Commit, error) {
	query := `
		SELECT id, project_id, branch_id, parent_commit_id, author, message, created_at
		FROM commits
		WHERE id = $1
	`

	var commit models.Commit
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&commit.ID,
		&commit.ProjectID,
		&commit.BranchID,
		&commit.ParentCommitID,
		&commit.Author,
		&commit.Message,
		&commit.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("commit not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	return &commit, nil
}

func (r *CommitRepository) ListByBranch(ctx context.Context, branchID uuid.UUID, limit, offset int) ([]models.Commit, error) {
	query := `
		SELECT id, project_id, branch_id, parent_commit_id, author, message, created_at
		FROM commits
		WHERE branch_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, branchID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
	}
	defer rows.Close()

	var commits []models.Commit
	for rows.Next() {
		var c models.Commit
		err := rows.Scan(
			&c.ID,
			&c.ProjectID,
			&c.BranchID,
			&c.ParentCommitID,
			&c.Author,
			&c.Message,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan commit: %w", err)
		}
		commits = append(commits, c)
	}

	return commits, nil
}
