package handler

import (
	"net/http"
)

type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

type HealthHandler struct {
	buildInfo BuildInfo
}

func NewHealthHandler(info BuildInfo) *HealthHandler {
	return &HealthHandler{buildInfo: info}
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "UP",
	})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "READY",
	})
}

func (h *HealthHandler) Version(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, h.buildInfo)
}
