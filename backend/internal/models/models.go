package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Branch struct {
	ID           uuid.UUID  `json:"id"`
	ProjectID    uuid.UUID  `json:"project_id"`
	Name         string     `json:"name"`
	HeadCommitID *uuid.UUID `json:"head_commit_id"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Commit struct {
	ID             uuid.UUID     `json:"id"`
	ProjectID      uuid.UUID     `json:"project_id"`
	BranchID       uuid.UUID     `json:"branch_id"`
	ParentCommitID *uuid.UUID    `json:"parent_commit_id"`
	Author         string        `json:"author"`
	Message        string        `json:"message"`
	CreatedAt      time.Time     `json:"created_at"`
	FileVersions   []FileVersion `json:"file_versions,omitempty"`
}

type File struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
}

type FileVersion struct {
	ID          uuid.UUID `json:"id"`
	FileID      uuid.UUID `json:"file_id"`
	CommitID    uuid.UUID `json:"commit_id"`
	StoragePath string    `json:"storage_path"`
	FileSize    int64     `json:"file_size"`
	Checksum    string    `json:"checksum"`
	CreatedAt   time.Time `json:"created_at"`
	Filename    string    `json:"filename,omitempty"`
}

type MergeRequest struct {
	ID             uuid.UUID  `json:"id"`
	ProjectID      uuid.UUID  `json:"project_id"`
	SourceBranchID uuid.UUID  `json:"source_branch_id"`
	TargetBranchID uuid.UUID  `json:"target_branch_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Author         string     `json:"author"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	MergedAt       *time.Time `json:"merged_at"`
}

type MergeConflict struct {
	ID              uuid.UUID  `json:"id"`
	MergeRequestID  uuid.UUID  `json:"merge_request_id"`
	FileID          uuid.UUID  `json:"file_id"`
	SourceVersionID uuid.UUID  `json:"source_version_id"`
	TargetVersionID uuid.UUID  `json:"target_version_id"`
	Status          string     `json:"status"`
	ResolutionNotes string     `json:"resolution_notes"`
	ResolvedAt      *time.Time `json:"resolved_at"`
}

type Comment struct {
	ID             uuid.UUID `json:"id"`
	MergeRequestID uuid.UUID `json:"merge_request_id"`
	Author         string    `json:"author"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type Approval struct {
	ID             uuid.UUID `json:"id"`
	MergeRequestID uuid.UUID `json:"merge_request_id"`
	Approver       string    `json:"approver"`
	CreatedAt      time.Time `json:"created_at"`
}
