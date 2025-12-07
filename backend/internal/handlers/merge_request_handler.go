package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rhblitstein/cad-version-control/internal/models"
	"github.com/rhblitstein/cad-version-control/internal/repository"
	"github.com/rhblitstein/cad-version-control/pkg/utils"
)

type MergeRequestHandler struct {
	mrRepo     *repository.MergeRequestRepository
	branchRepo *repository.BranchRepository
	fileRepo   *repository.FileRepository
}

func NewMergeRequestHandler(
	mrRepo *repository.MergeRequestRepository,
	branchRepo *repository.BranchRepository,
	fileRepo *repository.FileRepository,
) *MergeRequestHandler {
	return &MergeRequestHandler{
		mrRepo:     mrRepo,
		branchRepo: branchRepo,
		fileRepo:   fileRepo,
	}
}

func (h *MergeRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID      uuid.UUID `json:"project_id"`
		SourceBranchID uuid.UUID `json:"source_branch_id"`
		TargetBranchID uuid.UUID `json:"target_branch_id"`
		Title          string    `json:"title"`
		Description    string    `json:"description"`
		Author         string    `json:"author"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" || req.Author == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Title and author are required")
		return
	}

	// Get source and target branches
	sourceBranch, err := h.branchRepo.GetByID(r.Context(), req.SourceBranchID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Source branch not found")
		return
	}

	targetBranch, err := h.branchRepo.GetByID(r.Context(), req.TargetBranchID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Target branch not found")
		return
	}

	mr := &models.MergeRequest{
		ProjectID:      req.ProjectID,
		SourceBranchID: req.SourceBranchID,
		TargetBranchID: req.TargetBranchID,
		Title:          req.Title,
		Description:    req.Description,
		Author:         req.Author,
		Status:         "open",
	}

	if err := h.mrRepo.Create(r.Context(), mr); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create merge request")
		return
	}

	// Detect conflicts
	conflicts, err := h.detectConflicts(r.Context(), sourceBranch, targetBranch, mr.ID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to detect conflicts")
		return
	}

	// Create conflict records
	for _, conflict := range conflicts {
		if err := h.mrRepo.CreateConflict(r.Context(), &conflict); err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create conflict record")
			return
		}
	}

	utils.JSONResponse(w, http.StatusCreated, map[string]interface{}{
		"merge_request": mr,
		"conflicts":     conflicts,
	})
}

func (h *MergeRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	var projectID *uuid.UUID
	if projectIDStr != "" {
		parsed, err := uuid.Parse(projectIDStr)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid project ID")
			return
		}
		projectID = &parsed
	}

	mrs, err := h.mrRepo.List(r.Context(), projectID, status)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list merge requests")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"merge_requests": mrs,
	})
}

func (h *MergeRequestHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid merge request ID")
		return
	}

	mr, err := h.mrRepo.GetByID(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Merge request not found")
		return
	}

	// Get conflicts
	conflicts, err := h.mrRepo.GetConflicts(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get conflicts")
		return
	}

	// Get comments
	comments, err := h.mrRepo.GetComments(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get comments")
		return
	}

	// Get approvals
	approvals, err := h.mrRepo.GetApprovals(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get approvals")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"merge_request": mr,
		"conflicts":     conflicts,
		"comments":      comments,
		"approvals":     approvals,
	})
}

func (h *MergeRequestHandler) Approve(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid merge request ID")
		return
	}

	var req struct {
		Approver string `json:"approver"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	approval := &models.Approval{
		MergeRequestID: id,
		Approver:       req.Approver,
	}

	if err := h.mrRepo.CreateApproval(r.Context(), approval); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create approval")
		return
	}

	// Update MR status to approved
	if err := h.mrRepo.UpdateStatus(r.Context(), id, "approved"); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to update status")
		return
	}

	utils.JSONResponse(w, http.StatusOK, approval)
}

func (h *MergeRequestHandler) Merge(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid merge request ID")
		return
	}

	// Check for unresolved conflicts
	conflicts, err := h.mrRepo.GetConflicts(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get conflicts")
		return
	}

	for _, conflict := range conflicts {
		if conflict.Status == "unresolved" {
			utils.ErrorResponse(w, http.StatusBadRequest, "Cannot merge: unresolved conflicts exist")
			return
		}
	}

	// TODO: Create merge commit (simplified for now)
	if err := h.mrRepo.UpdateStatus(r.Context(), id, "merged"); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to update status")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Merge request merged successfully",
	})
}

func (h *MergeRequestHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid merge request ID")
		return
	}

	var req struct {
		Author  string `json:"author"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	comment := &models.Comment{
		MergeRequestID: id,
		Author:         req.Author,
		Content:        req.Content,
	}

	if err := h.mrRepo.CreateComment(r.Context(), comment); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	utils.JSONResponse(w, http.StatusCreated, comment)
}

func (h *MergeRequestHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid merge request ID")
		return
	}

	comments, err := h.mrRepo.GetComments(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get comments")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"comments": comments,
	})
}

func (h *MergeRequestHandler) GetConflicts(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid merge request ID")
		return
	}

	conflicts, err := h.mrRepo.GetConflicts(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get conflicts")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"conflicts": conflicts,
	})
}

func (h *MergeRequestHandler) ResolveConflict(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid conflict ID")
		return
	}

	var req struct {
		ResolutionNotes string    `json:"resolution_notes"`
		ChosenVersionID uuid.UUID `json:"chosen_version_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.mrRepo.ResolveConflict(r.Context(), id, req.ResolutionNotes); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to resolve conflict")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Conflict resolved",
	})
}

func (h *MergeRequestHandler) GetDiff(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid conflict ID")
		return
	}

	conflict, err := h.mrRepo.GetConflictByID(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Conflict not found")
		return
	}

	sourceVersion, err := h.fileRepo.GetVersionByID(r.Context(), conflict.SourceVersionID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get source version")
		return
	}

	targetVersion, err := h.fileRepo.GetVersionByID(r.Context(), conflict.TargetVersionID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get target version")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"conflict_id": id,
		"source_version": map[string]interface{}{
			"id":           sourceVersion.ID,
			"filename":     sourceVersion.Filename,
			"download_url": "/api/file-versions/" + sourceVersion.ID.String() + "/download",
			"file_size":    sourceVersion.FileSize,
		},
		"target_version": map[string]interface{}{
			"id":           targetVersion.ID,
			"filename":     targetVersion.Filename,
			"download_url": "/api/file-versions/" + targetVersion.ID.String() + "/download",
			"file_size":    targetVersion.FileSize,
		},
		"diff_summary": map[string]interface{}{
			"geometry_changed": sourceVersion.Checksum != targetVersion.Checksum,
			"size_diff":        targetVersion.FileSize - sourceVersion.FileSize,
		},
	})
}

// Helper function to detect conflicts
func (h *MergeRequestHandler) detectConflicts(ctx context.Context, sourceBranch, targetBranch *models.Branch, mrID uuid.UUID) ([]models.MergeConflict, error) {
	var conflicts []models.MergeConflict

	// Get files from both branches
	if sourceBranch.HeadCommitID == nil || targetBranch.HeadCommitID == nil {
		return conflicts, nil // No commits yet
	}

	sourceFiles, err := h.fileRepo.GetVersionsByCommit(ctx, *sourceBranch.HeadCommitID)
	if err != nil {
		return nil, err
	}

	targetFiles, err := h.fileRepo.GetVersionsByCommit(ctx, *targetBranch.HeadCommitID)
	if err != nil {
		return nil, err
	}

	// Create map of target files by filename
	targetFileMap := make(map[string]models.FileVersion)
	for _, tf := range targetFiles {
		targetFileMap[tf.Filename] = tf
	}

	// Check for conflicts
	for _, sf := range sourceFiles {
		if tf, exists := targetFileMap[sf.Filename]; exists {
			// File exists in both branches
			if sf.Checksum != tf.Checksum {
				// Different checksums = conflict
				conflicts = append(conflicts, models.MergeConflict{
					MergeRequestID:  mrID,
					FileID:          sf.FileID,
					SourceVersionID: sf.ID,
					TargetVersionID: tf.ID,
					Status:          "unresolved",
				})
			}
		}
	}

	return conflicts, nil
}
