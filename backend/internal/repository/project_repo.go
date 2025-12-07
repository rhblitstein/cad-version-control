package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/rhblitstein/cad-version-control/internal/models"
)

type ProjectRepository struct {
	db    *sql.DB
	cache *RedisClient
}

func NewProjectRepository(db *sql.DB, cache *RedisClient) *ProjectRepository {
	return &ProjectRepository{
		db:    db,
		cache: cache,
	}
}

func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	query := `
		INSERT INTO projects (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	project.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		project.ID,
		project.Name,
		project.Description,
	).Scan(&project.CreatedAt, &project.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	var project models.Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]models.Project, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}
