package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kostayne/go-microservice/services/delivery/internal/model"
	"github.com/kostayne/go-microservice/services/delivery/internal/repository"
)

type Handler struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/health", h.health)
	r.Get("/couriers", h.listCouriers)
	return r
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listCouriers(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListCouriers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list couriers")
		return
	}
	if list == nil {
		list = []model.Courier{}
	}
	writeJSON(w, http.StatusOK, list)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
