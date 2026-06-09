package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kostayne/go-microservice/pkg/auth"
	"github.com/kostayne/go-microservice/services/user/internal/model"
	"github.com/kostayne/go-microservice/services/user/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo      *repository.Repository
	jwtSecret string
}

func New(repo *repository.Repository, jwtSecret string) *Handler {
	return &Handler{repo: repo, jwtSecret: jwtSecret}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/health", h.health)
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Get("/users/{id}", h.getUser)
	r.Get("/users/{id}/addresses", h.listAddresses)
	r.Post("/users/{id}/addresses", h.addAddress)
	r.Get("/internal/validate/{id}", h.validateUser)
	return r
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "email, password and name are required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash password")
		return
	}

	user, err := h.repo.CreateUser(r.Context(), req.Email, string(hash), req.Name, req.Phone)
	if err != nil {
		writeError(w, http.StatusConflict, "user already exists")
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID, user.Email, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "generate token")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":  user,
		"token": token,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	user, err := h.repo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID, user.Email, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) validateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"valid": true,
	})
}

func (h *Handler) listAddresses(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	addrs, err := h.repo.ListAddresses(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list addresses")
		return
	}
	if addrs == nil {
		addrs = []model.Address{}
	}
	writeJSON(w, http.StatusOK, addrs)
}

func (h *Handler) addAddress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var addr model.Address
	if err := json.NewDecoder(r.Body).Decode(&addr); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	addr.UserID = id
	if err := h.repo.AddAddress(r.Context(), &addr); err != nil {
		writeError(w, http.StatusInternalServerError, "add address")
		return
	}
	writeJSON(w, http.StatusCreated, addr)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
