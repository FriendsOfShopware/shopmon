package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/uptime"
)

// UpdateUptimeSettings creates or updates the uptime monitor for an environment.
func (h *Handler) UpdateUptimeSettings(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req api.UptimeSettingsRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	settings := uptime.Settings{
		Enabled:           req.Enabled,
		IntervalSeconds:   int32(req.IntervalSeconds),
		ExpectedStatus:    int32(req.ExpectedStatus),
		FailureThreshold:  int32(req.FailureThreshold),
		RecoveryThreshold: int32(req.RecoveryThreshold),
	}
	if req.Url != nil {
		settings.URL = *req.Url
	}
	if req.ContentMatch != nil {
		settings.ContentMatch = *req.ContentMatch
	}

	err := h.uptime.UpdateSettings(r.Context(), user.ID, int32(environmentId), settings)
	if h.writeUptimeError(w, r, "update uptime settings", err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetEnvironmentUptime returns uptime status, availability and incident history.
func (h *Handler) GetEnvironmentUptime(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId, params api.GetEnvironmentUptimeParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rangeKey := "24h"
	if params.Range != nil {
		rangeKey = string(*params.Range)
	}

	view, err := h.uptime.GetView(r.Context(), user.ID, int32(environmentId), rangeKey)
	if h.writeUptimeError(w, r, "get uptime", err) {
		return
	}

	httputil.WriteJSON(w, http.StatusOK, uptimeViewToResponse(view))
}

func (h *Handler) writeUptimeError(w http.ResponseWriter, r *http.Request, operation string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, uptime.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
	case errors.Is(err, uptime.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
	case errors.Is(err, uptime.ErrInvalidSettings):
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		slog.ErrorContext(r.Context(), "uptime operation failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to "+operation)
	}
	return true
}

func uptimeViewToResponse(view uptime.View) api.UptimeResponse {
	resp := api.UptimeResponse{
		Settings:     uptimeSettingsToAPI(view.Settings),
		Availability: view.Availability,
		Days:         make([]api.UptimeDay, 0, len(view.Days)),
		Incidents:    make([]api.UptimeIncident, 0, len(view.Incidents)),
		Latency:      make([]api.UptimeLatencyPoint, 0, len(view.Latency)),
	}

	if view.OpenIncident != nil {
		incident := uptimeIncidentToAPI(*view.OpenIncident)
		resp.OpenIncident = &incident
	}

	for _, d := range view.Days {
		resp.Days = append(resp.Days, api.UptimeDay{
			Day:          d.Day,
			Availability: d.Availability,
			DownSeconds:  d.DownSeconds,
		})
	}
	for _, i := range view.Incidents {
		resp.Incidents = append(resp.Incidents, uptimeIncidentToAPI(i))
	}
	for _, l := range view.Latency {
		resp.Latency = append(resp.Latency, api.UptimeLatencyPoint{
			Timestamp: l.Timestamp,
			AvgMs:     int32PtrToIntPtr(l.AvgMs),
			P95Ms:     int32PtrToIntPtr(l.P95Ms),
			ProbesRun: int(l.ProbesRun),
		})
	}
	return resp
}

func uptimeSettingsToAPI(s uptime.SettingsView) api.UptimeSettings {
	out := api.UptimeSettings{
		Enabled:           s.Enabled,
		Url:               s.URL,
		IntervalSeconds:   int(s.IntervalSeconds),
		ExpectedStatus:    int(s.ExpectedStatus),
		ContentMatch:      s.ContentMatch,
		FailureThreshold:  int(s.FailureThreshold),
		RecoveryThreshold: int(s.RecoveryThreshold),
		Status:            api.UptimeSettingsStatus(s.Status),
		LastCheckedAt:     s.LastCheckedAt,
		LastStatusCode:    int32PtrToIntPtr(s.LastStatusCode),
		LastLatencyMs:     int32PtrToIntPtr(s.LastLatencyMs),
		LastError:         s.LastError,
	}
	return out
}

func uptimeIncidentToAPI(i uptime.Incident) api.UptimeIncident {
	return api.UptimeIncident{
		Id:              int(i.ID),
		StartedAt:       i.StartedAt,
		ResolvedAt:      i.ResolvedAt,
		DurationSeconds: i.DurationSeconds,
		StatusCode:      int32PtrToIntPtr(i.StatusCode),
		LatencyMs:       int32PtrToIntPtr(i.LatencyMs),
		Error:           i.Error,
	}
}

func int32PtrToIntPtr(v *int32) *int {
	if v == nil {
		return nil
	}
	out := int(*v)
	return &out
}
