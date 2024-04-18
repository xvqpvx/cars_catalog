package app

import (
	"car_catalog/internal/config"
	"car_catalog/internal/database"
	"car_catalog/internal/handler"
	"car_catalog/internal/repository"
	"car_catalog/internal/router"
	"car_catalog/internal/service"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	Server *http.Server
}

func New() (*App, error) {
	log.Println("[INFO] Creating new application instance")

	cfg := config.New()

	err := database.MigrateDatabase(cfg)
	if err != nil {
		log.Printf("[ERROR] Failed to migrate database: %v", err)
		return nil, err
	}

	conn := database.DatabaseConnection(cfg)

	carRepo := repository.NewCarRepository(conn)
	carService := service.NewCarService(carRepo)
	carHandler := handler.NewCarHandler(carService, cfg.ExternalAPIURL)

	routes := router.NewRouter(carHandler)

	server := &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort), Handler: routes}

	log.Println("[INFO] Application instance created successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("[INFO] Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("[ERROR] Server shutdown failed: %v", err)
		}

		log.Println("[INFO] Server shutdown completed")
	}()

	return &App{Server: server}, nil
}

func (a *App) Run() error {
	log.Printf("[INFO] Starting server on %s", a.Server.Addr)
	err := a.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Printf("[ERROR] Server stopped with error: %v", err)
	} else {
		log.Println("[INFO] Server stopped gracefully")
	}
	return err
}
