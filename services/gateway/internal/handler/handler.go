package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kostayne/go-microservice/pkg/auth"
	"github.com/kostayne/go-microservice/services/gateway/internal/docs"
	"github.com/kostayne/go-microservice/services/gateway/internal/proxy"
)

type Config struct {
	UserURL       string
	RestaurantURL string
	OrderURL      string
	PaymentURL    string
	DeliveryURL   string
	JWTSecret     string
}

type Handler struct {
	cfg    Config
	client *proxy.Client
}

func New(cfg Config) *Handler {
	return &Handler{cfg: cfg, client: proxy.New()}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	docs.Register(r)

	r.Get("/health", h.health)

	r.Post("/api/auth/register", h.proxyTo(h.cfg.UserURL+"/register"))
	r.Post("/api/auth/login", h.proxyTo(h.cfg.UserURL+"/login"))

	r.Get("/api/restaurants", h.proxyTo(h.cfg.RestaurantURL+"/restaurants"))
	r.Get("/api/restaurants/{id}", h.proxyTo(h.cfg.RestaurantURL+"/restaurants/{id}"))
	r.Get("/api/restaurants/{id}/menu", h.proxyTo(h.cfg.RestaurantURL+"/restaurants/{id}/menu"))
	r.Get("/api/couriers", h.proxyTo(h.cfg.DeliveryURL+"/couriers"))

	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Post("/api/orders", h.createOrder)
		r.Get("/api/orders/{id}", h.getOrderWithDetails)
		r.Get("/api/users/me/orders", h.myOrders)
		r.Get("/api/users/me", h.me)
	})

	return r
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "api-gateway"})
}

func (h *Handler) proxyTo(targetPattern string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := targetPattern
		if id := chi.URLParam(r, "id"); id != "" {
			target = strings.Replace(target, "{id}", id, 1)
		}
		h.client.Forward(w, r, target)
	}
}

type ctxKey string

const userIDKey ctxKey = "user_id"

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ParseToken(h.cfg.JWTSecret, token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(string)
	var user any
	if err := h.client.Get(h.cfg.UserURL+"/users/"+userID, &user); err != nil {
		writeError(w, http.StatusBadGateway, "fetch user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type createOrderBody struct {
	RestaurantID string `json:"restaurant_id"`
	Items        []struct {
		MenuItemID string `json:"menu_item_id"`
		Quantity   int    `json:"quantity"`
	} `json:"items"`
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(string)
	var body createOrderBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	payload := map[string]any{
		"user_id":       userID,
		"restaurant_id": body.RestaurantID,
		"items":         body.Items,
	}

	var order any
	if err := h.client.Post(h.cfg.OrderURL+"/orders", payload, &order); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) getOrderWithDetails(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	var order map[string]any
	if err := h.client.Get(h.cfg.OrderURL+"/orders/"+orderID, &order); err != nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	var payment any
	if err := h.client.Get(h.cfg.PaymentURL+"/payments/order/"+orderID, &payment); err == nil {
		order["payment"] = payment
	}

	writeJSON(w, http.StatusOK, order)
}

func (h *Handler) myOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(string)
	var orders any
	if err := h.client.Get(h.cfg.OrderURL+"/users/"+userID+"/orders", &orders); err != nil {
		writeError(w, http.StatusBadGateway, "fetch orders")
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
