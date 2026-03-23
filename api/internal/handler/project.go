package handler

import (
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// GetOrganizationProjects returns all projects in an organization.
func (h *Handler) GetOrganizationProjects(w http.ResponseWriter, r *http.Request, orgId api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	rows, err := h.queries.ListProjectsByOrganization(r.Context(), orgId)
	if err != nil {
		slog.Error("failed to list projects", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get projects")
		return
	}

	result := make([]api.Project, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.Project{
			Id:             int(row.ID),
			Name:           row.Name,
			Description:    row.Description,
			GitUrl:         row.GitUrl,
			OrganizationId: row.OrganizationID,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateProject creates a new project in an organization.
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request, orgId api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	var req api.CreateProjectRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	projectID, err := h.queries.CreateProject(r.Context(), queries.CreateProjectParams{
		OrganizationID: orgId,
		Name:           req.Name,
		Description:    req.Description,
		GitUrl:         req.GitUrl,
	})
	if err != nil {
		slog.Error("failed to create project", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, api.Project{
		Id:             int(projectID),
		Name:           req.Name,
		Description:    req.Description,
		GitUrl:         req.GitUrl,
		OrganizationId: orgId,
	})
}

// UpdateProject updates an existing project.
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request, orgId api.OrgId, projectId api.ProjectId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	project, err := h.queries.GetProjectByID(r.Context(), int32(projectId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "project not found")
		return
	}

	if project.OrganizationID != orgId {
		httputil.WriteError(w, http.StatusForbidden, "project does not belong to this organization")
		return
	}

	var req api.UpdateProjectRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	name := project.Name
	if req.Name != nil {
		name = *req.Name
	}
	description := project.Description
	if req.Description != nil {
		description = req.Description
	}
	gitUrl := project.GitUrl
	if req.GitUrl != nil {
		gitUrl = req.GitUrl
	}

	if err := h.queries.UpdateProject(r.Context(), queries.UpdateProjectParams{
		Name:           name,
		Description:    description,
		GitUrl:         gitUrl,
		ID:             int32(projectId),
		OrganizationID: orgId,
	}); err != nil {
		slog.Error("failed to update project", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update project")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteProject deletes a project.
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request, orgId api.OrgId, projectId api.ProjectId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	// Check that the project belongs to this organization
	project, err := h.queries.GetProjectByID(r.Context(), int32(projectId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "project not found")
		return
	}

	if project.OrganizationID != orgId {
		httputil.WriteError(w, http.StatusForbidden, "project does not belong to this organization")
		return
	}

	shopCount, err := h.queries.CountShopsInProject(r.Context(), int32(projectId))
	if err == nil && shopCount > 0 {
		httputil.WriteError(w, http.StatusConflict, "cannot delete project with existing shops")
		return
	}

	if err := h.queries.DeleteProject(r.Context(), queries.DeleteProjectParams{
		ID:             int32(projectId),
		OrganizationID: orgId,
	}); err != nil {
		slog.Error("failed to delete project", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
