package main

import (
	"log/slog"
	"net/http"
	"os"

	_ "example.com/order-service/docs"
	"example.com/order-service/internal/handler"
	"example.com/order-service/internal/repository/memory"
	"example.com/order-service/internal/service"
)

// @title           Order Service API
// @version         1.0
// @description     Implements order creation, retrieval, status update and history
// @host            127.0.0.1:8080
// @BasePath        /api/v1
// @schemes         http

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	repository := memory.NewOrderRepository()
	orderService := service.NewOrderService(repository)
	router := handler.NewRouter(orderService, logger)

	address := envOrDefault("HTTP_ADDR", ":8080")
	logger.Info("starting HTTP server", slog.String("address", address))

	server := &http.Server{
		Addr:    address,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
