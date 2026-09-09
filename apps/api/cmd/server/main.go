package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kristopher1027/devflow-ai/internal/config"
	devflowhttp "github.com/kristopher1027/devflow-ai/internal/http"
	"github.com/kristopher1027/devflow-ai/internal/server"
)

func main() {
	cfg := config.Load()

	router := devflowhttp.NewRouter()
	app := server.New(":"+cfg.Port, router)

	go func() {
		log.Printf("DevFlow API running on http://localhost:%s", cfg.Port)

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
