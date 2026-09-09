package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/config"
	"github.com/kristopher1027/devflow-ai/internal/database"
	devflowhttp "github.com/kristopher1027/devflow-ai/internal/http"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/server"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	log.Println("database connection established")

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := devflowhttp.NewUserHandler(userService)

	router := devflowhttp.NewRouter(userHandler)

	app := server.New(":"+cfg.Port, router)

	go func() {
		log.Printf(
			"DevFlow API running on http://localhost:%s",
			cfg.Port,
		)

		if err := app.Start(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("shutting down DevFlow API...")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	log.Println("DevFlow API stopped")
}
