package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/cors"
	"github.com/kostayne/go-microservice/services/gateway/internal/handler"
)

func main() {
	port := env("PORT", "8080")
	jwtSecret := env("JWT_SECRET", "dev-secret-change-me")

	cfg := handler.Config{
		UserURL:       env("USER_SVC_URL", "http://localhost:8081"),
		RestaurantURL: env("RESTAURANT_SVC_URL", "http://localhost:8082"),
		OrderURL:      env("ORDER_SVC_URL", "http://localhost:8083"),
		PaymentURL:    env("PAYMENT_SVC_URL", "http://localhost:8084"),
		DeliveryURL:   env("DELIVERY_SVC_URL", "http://localhost:8085"),
		JWTSecret:     jwtSecret,
	}

	h := handler.New(cfg)
	corsHandler := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      corsHandler(h.Routes()),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("api-gateway listening on :%s", port)
	log.Printf("openapi spec: http://localhost:%s/openapi.yaml", port)
	log.Printf("scalar docs:  http://localhost:%s/docs", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
