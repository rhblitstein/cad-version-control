package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rhblitstein/cad-version-control/internal/models"
	"github.com/rhblitstein/cad-version-control/internal/repository"
	"github.com/rhblitstein/cad-version-control/internal/storage"
	"github.com/rhblitstein/cad-version-control/pkg/utils"
)

type CommitHandler struct {
	commitRepo *repository.CommitRepository
	branchRepo *repository.BranchRepository
	fileRepo   *repository.FileRepository
	storage    *storage.MinioClient
}

func NewCommitHandler(
	commitRepo *repository.CommitRepository,
	branchRepo *repository.BranchRepository,
	fileRepo *repository.FileRepository,
	storage *storage.MinioClient,
) *CommitHandler {
	return &CommitHandler{
		commitRepo: commitRepo,
		branchRepo: branchRepo,
		fileRepo:   fileRepo,
		storage:    storage,
	}
}

func (h *CommitHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	branchIDStr := r.FormValue("branch_id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid branch ID")
		return
	}

	message := r.FormValue("message")
	author := r.FormValue("author")

	if message == "" || author == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Message and author are required")
		return
	}

	branch, err := h.branchRepo.GetByID(r.Context(), branchID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Branch not found")
		return
	}

	commit := &models.Commit{
		ProjectID:      projectID,
		BranchID:       branchID,
		ParentCommitID: branch.HeadCommitID,
		Author:         author,
		Message:        message,
	}

	if err := h.commitRepo.Create(r.Context(), commit); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create commit")
		return
	}

	files := r.MultipartForm.File["files"]
	var fileVersions []models.FileVersion

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to read file")
			return
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to read file content")
			return
		}

		hash := sha256.Sum256(content)
		checksum := hex.EncodeToString(hash[:])

		existingFile, err := h.fileRepo.GetByFilename(r.Context(), projectID, fileHeader.Filename)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to check file")
			return
		}

		var fileID uuid.UUID
		if existingFile == nil {
			newFile := &models.File{
				ProjectID: projectID,
				Filename:  fileHeader.Filename,
			}
			if err := h.fileRepo.Create(r.Context(), newFile); err != nil {
				utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create file")
				return
			}
			fileID = newFile.ID
		} else {
			fileID = existingFile.ID
		}

		existingVersion, err := h.fileRepo.ChecksumExists(r.Context(), checksum)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to check checksum")
			return
		}

		var storagePath string
		if existingVersion != nil {
			storagePath = existingVersion.StoragePath
		} else {
			storagePath = fmt.Sprintf("projects/%s/commits/%s/%s", projectID, commit.ID, fileID)

			file.Seek(0, 0)

			if err := h.storage.Upload(r.Context(), storagePath, file, fileHeader.Size, "application/octet-stream"); err != nil {
				utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to upload file")
				return
			}
		}

		version := &models.FileVersion{
			FileID:      fileID,
			CommitID:    commit.ID,
			StoragePath: storagePath,
			FileSize:    fileHeader.Size,
			Checksum:    checksum,
		}

		if err := h.fileRepo.CreateVersion(r.Context(), version); err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create file version")
			return
		}

		version.Filename = fileHeader.Filename
		fileVersions = append(fileVersions, *version)
	}

	if err := h.branchRepo.UpdateHead(r.Context(), branchID, commit.ID); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to update branch")
		return
	}

	commit.FileVersions = fileVersions

	utils.JSONResponse(w, http.StatusCreated, commit)
}

func (h *CommitHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid commit ID")
		return
	}

	commit, err := h.commitRepo.GetByID(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "Commit not found")
		return
	}

	fileVersions, err := h.fileRepo.GetVersionsByCommit(r.Context(), id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get file versions")
		return
	}

	commit.FileVersions = fileVersions

	utils.JSONResponse(w, http.StatusOK, commit)
}

func (h *CommitHandler) ListByBranch(w http.ResponseWriter, r *http.Request) {
	branchIDStr := chi.URLParam(r, "branch_id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid branch ID")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 {
		limit = 50
	}

	commits, err := h.commitRepo.ListByBranch(r.Context(), branchID, limit, offset)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to list commits")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"commits": commits,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *CommitHandler) GetFileVersions(w http.ResponseWriter, r *http.Request) {
	utils.ErrorResponse(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *CommitHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	versionIDStr := chi.URLParam(r, "id")
	versionID, err := uuid.Parse(versionIDStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid version ID")
		return
	}

	version, err := h.fileRepo.GetVersionByID(r.Context(), versionID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "File version not found")
		return
	}

	object, err := h.storage.Download(r.Context(), version.StoragePath)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to download file")
		return
	}
	defer object.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", version.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", version.FileSize))

	if _, err := io.Copy(w, object); err != nil {
		return
	}
}
