package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kostayne/go-microservice/services/restaurant/internal/model"
	"github.com/kostayne/go-microservice/services/restaurant/internal/repository"
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
	r.Get("/restaurants", h.listRestaurants)
	r.Get("/restaurants/{id}", h.getRestaurant)
	r.Get("/restaurants/{id}/menu", h.getMenu)
	r.Post("/internal/menu/validate", h.validateMenu)
	return r
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listRestaurants(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListRestaurants(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list restaurants")
		return
	}
	if list == nil {
		list = []model.Restaurant{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) getRestaurant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rest, err := h.repo.GetRestaurant(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "restaurant not found")
		return
	}
	writeJSON(w, http.StatusOK, rest)
}

func (h *Handler) getMenu(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	items, err := h.repo.ListMenu(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list menu")
		return
	}
	if items == nil {
		items = []model.MenuItem{}
	}
	writeJSON(w, http.StatusOK, items)
}

type validateMenuRequest struct {
	RestaurantID string   `json:"restaurant_id"`
	ItemIDs      []string `json:"item_ids"`
}

func (h *Handler) validateMenu(w http.ResponseWriter, r *http.Request) {
	var req validateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	rest, err := h.repo.GetRestaurant(r.Context(), req.RestaurantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "restaurant not found")
		return
	}
	if !rest.IsOpen {
		writeError(w, http.StatusBadRequest, "restaurant is closed")
		return
	}

	items, err := h.repo.GetMenuItems(r.Context(), req.RestaurantID, req.ItemIDs)
	if err != nil || len(items) != len(req.ItemIDs) {
		writeError(w, http.StatusBadRequest, "invalid menu items")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
