package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/cors"
	"github.com/kostayne/go-microservice/pkg/config"
	"github.com/kostayne/go-microservice/pkg/telemetry"
	"github.com/kostayne/go-microservice/services/gateway/internal/handler"
)

const serviceName = "api-gateway"

func main() {
	log.Printf("api-gateway starting (APP_ENV=%s)", config.AppEnv())

	shutdown, err := telemetry.Init(context.Background(), serviceName)
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	port := config.String("PORT", "8080")
	jwtSecret := config.String("JWT_SECRET", "dev-secret-change-me")

	cfg := handler.Config{
		UserURL:       config.String("USER_SVC_URL", "http://localhost:8081"),
		RestaurantURL: config.String("RESTAURANT_SVC_URL", "http://localhost:8082"),
		OrderURL:      config.String("ORDER_SVC_URL", "http://localhost:8083"),
		PaymentURL:    config.String("PAYMENT_SVC_URL", "http://localhost:8084"),
		DeliveryURL:   config.String("DELIVERY_SVC_URL", "http://localhost:8085"),
		JWTSecret:     jwtSecret,
	}

	h := handler.New(cfg)
	origins := config.CORSOrigins(true)
	corsHandler := cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: !containsWildcard(origins),
		MaxAge:           300,
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      telemetry.WrapHTTP(serviceName, corsHandler(h.Routes())),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("api-gateway listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}

func containsWildcard(origins []string) bool {
	for _, o := range origins {
		if o == "*" {
			return true
		}
	}
	return false
}
