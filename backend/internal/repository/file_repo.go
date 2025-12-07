package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/rhblitstein/cad-version-control/internal/models"
)

type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (id, project_id, filename, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING created_at
	`

	file.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		file.ID,
		file.ProjectID,
		file.Filename,
	).Scan(&file.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

func (r *FileRepository) GetByFilename(ctx context.Context, projectID uuid.UUID, filename string) (*models.File, error) {
	query := `
		SELECT id, project_id, filename, created_at
		FROM files
		WHERE project_id = $1 AND filename = $2
	`

	var file models.File
	err := r.db.QueryRowContext(ctx, query, projectID, filename).Scan(
		&file.ID,
		&file.ProjectID,
		&file.Filename,
		&file.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not an error, just doesn't exist
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return &file, nil
}

func (r *FileRepository) CreateVersion(ctx context.Context, version *models.FileVersion) error {
	query := `
		INSERT INTO file_versions (id, file_id, commit_id, storage_path, file_size, checksum, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING created_at
	`

	version.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		version.ID,
		version.FileID,
		version.CommitID,
		version.StoragePath,
		version.FileSize,
		version.Checksum,
	).Scan(&version.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create file version: %w", err)
	}

	return nil
}

func (r *FileRepository) GetVersionsByCommit(ctx context.Context, commitID uuid.UUID) ([]models.FileVersion, error) {
	query := `
		SELECT fv.id, fv.file_id, fv.commit_id, fv.storage_path, fv.file_size, fv.checksum, fv.created_at, f.filename
		FROM file_versions fv
		JOIN files f ON fv.file_id = f.id
		WHERE fv.commit_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, commitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file versions: %w", err)
	}
	defer rows.Close()

	var versions []models.FileVersion
	for rows.Next() {
		var v models.FileVersion
		err := rows.Scan(
			&v.ID,
			&v.FileID,
			&v.CommitID,
			&v.StoragePath,
			&v.FileSize,
			&v.Checksum,
			&v.CreatedAt,
			&v.Filename,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}

func (r *FileRepository) GetVersionByID(ctx context.Context, versionID uuid.UUID) (*models.FileVersion, error) {
	query := `
		SELECT fv.id, fv.file_id, fv.commit_id, fv.storage_path, fv.file_size, fv.checksum, fv.created_at, f.filename
		FROM file_versions fv
		JOIN files f ON fv.file_id = f.id
		WHERE fv.id = $1
	`

	var version models.FileVersion
	err := r.db.QueryRowContext(ctx, query, versionID).Scan(
		&version.ID,
		&version.FileID,
		&version.CommitID,
		&version.StoragePath,
		&version.FileSize,
		&version.Checksum,
		&version.CreatedAt,
		&version.Filename,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file version not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file version: %w", err)
	}

	return &version, nil
}

func (r *FileRepository) ChecksumExists(ctx context.Context, checksum string) (*models.FileVersion, error) {
	query := `
		SELECT id, file_id, commit_id, storage_path, file_size, checksum, created_at
		FROM file_versions
		WHERE checksum = $1
		LIMIT 1
	`

	var version models.FileVersion
	err := r.db.QueryRowContext(ctx, query, checksum).Scan(
		&version.ID,
		&version.FileID,
		&version.CommitID,
		&version.StoragePath,
		&version.FileSize,
		&version.Checksum,
		&version.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not an error, just doesn't exist
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check checksum: %w", err)
	}

	return &version, nil
}
