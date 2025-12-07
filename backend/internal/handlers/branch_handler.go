package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rhblitstein/cad-version-control/internal/models"
	"github.com/rhblitstein/cad-version-control/internal/repository"
	"github.com/rhblitstein/cad-version-control/pkg/utils"
)

type BranchHandler struct {
	branchRepo  *repository.BranchRepository
	projectRepo *repository.ProjectRepository
}

func NewBranchHandler(branchRepo *repository.BranchRepository, projectRepo *repository.ProjectRepository) *BranchHandler {
	return &BranchHandler{
		branchRepo:  branchRepo,
		projectRepo: projectRepo,
	}
}

func (h *BranchHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req struct {
		Name           string     `json:"name"`
		SourceBranchID *uuid.UUID `json:"source_branch_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}

	branch := &models.Branch{
		ProjectID: projectID,
		Name:      req.Name,
	}

	if req.SourceBranchID != nil {
		sourceBranch, err := h.branchRepo.GetByID(r.Context(), *req.SourceBranchID)
		if err != nil {
			utils.ErrorResponse(w, http.StatusNotFound, "Source branch not found")
			return
		}
		branch.HeadCommitID = sourceBranch.HeadCommitID
	}

	if err := h.branchRepo.Create(r.Context(), branch); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create branch")
		return
	}

	utils.JSONResponse(w, http.StatusCreated, branch)
}

func (h *BranchHandler) List(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	branches, err := h.branchRepo.ListByProject(r.Context(), projectID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list branches")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"branches": branches,
	})
}

func (h *BranchHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid branch ID")
		return
	}

	branch, err := h.branchRepo.GetByID(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Branch not found")
		return
	}

	utils.JSONResponse(w, http.StatusOK, branch)
}
