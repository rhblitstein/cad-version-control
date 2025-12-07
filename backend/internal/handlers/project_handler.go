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

type ProjectHandler struct {
	repo *repository.ProjectRepository
}

func NewProjectHandler(repo *repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}

	project := &models.Project{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.repo.Create(r.Context(), project); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	utils.JSONResponse(w, http.StatusCreated, project)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.repo.List(r.Context())
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list projects")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"projects": projects,
	})
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	project, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Project not found")
		return
	}

	utils.JSONResponse(w, http.StatusOK, project)
}
