-- CAD Version Control System - PostgreSQL Schema

-- Projects: Top-level container
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Branches: Git-like branching
CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    head_commit_id UUID, -- Points to latest commit, set after commits table
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- Commits: Version snapshots
CREATE TABLE commits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    parent_commit_id UUID REFERENCES commits(id), -- NULL for initial commit
    author VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Add foreign key to branches after commits table exists
ALTER TABLE branches ADD CONSTRAINT fk_head_commit 
    FOREIGN KEY (head_commit_id) REFERENCES commits(id) ON DELETE SET NULL;

-- Files: CAD file metadata
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- File Versions: Track file changes across commits
CREATE TABLE file_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    commit_id UUID NOT NULL REFERENCES commits(id) ON DELETE CASCADE,
    storage_path VARCHAR(512) NOT NULL, -- Path in MinIO
    file_size BIGINT NOT NULL,
    checksum VARCHAR(64) NOT NULL, -- SHA-256 for deduplication
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(file_id, commit_id)
);

-- Merge Requests: Branch merge workflow
CREATE TABLE merge_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    target_branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'open', -- open, approved, merged, closed
    author VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP,
    CHECK (status IN ('open', 'approved', 'merged', 'closed'))
);

-- Merge Conflicts: Track file conflicts in merge requests
CREATE TABLE merge_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merge_request_id UUID NOT NULL REFERENCES merge_requests(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    source_version_id UUID NOT NULL REFERENCES file_versions(id),
    target_version_id UUID NOT NULL REFERENCES file_versions(id),
    status VARCHAR(50) NOT NULL DEFAULT 'unresolved', -- unresolved, resolved
    resolution_notes TEXT,
    resolved_at TIMESTAMP,
    CHECK (status IN ('unresolved', 'resolved'))
);

-- Comments: Discussion threads on merge requests
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merge_request_id UUID NOT NULL REFERENCES merge_requests(id) ON DELETE CASCADE,
    author VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Approvals: Track who approved merge requests
CREATE TABLE approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merge_request_id UUID NOT NULL REFERENCES merge_requests(id) ON DELETE CASCADE,
    approver VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(merge_request_id, approver)
);

-- Indexes for performance
CREATE INDEX idx_branches_project ON branches(project_id);
CREATE INDEX idx_commits_project ON commits(project_id);
CREATE INDEX idx_commits_branch ON commits(branch_id);
CREATE INDEX idx_commits_parent ON commits(parent_commit_id);
CREATE INDEX idx_files_project ON files(project_id);
CREATE INDEX idx_file_versions_file ON file_versions(file_id);
CREATE INDEX idx_file_versions_commit ON file_versions(commit_id);
CREATE INDEX idx_file_versions_checksum ON file_versions(checksum);
CREATE INDEX idx_merge_requests_project ON merge_requests(project_id);
CREATE INDEX idx_merge_requests_status ON merge_requests(status);
CREATE INDEX idx_merge_conflicts_mr ON merge_conflicts(merge_request_id);
CREATE INDEX idx_comments_mr ON comments(merge_request_id);
CREATE INDEX idx_approvals_mr ON approvals(merge_request_id);