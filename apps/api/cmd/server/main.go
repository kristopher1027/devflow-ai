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
	githubintegration "github.com/kristopher1027/devflow-ai/internal/integration/github"
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

	// Repositories
	userRepository := repository.NewUserRepository(db)
	sessionRepository := repository.NewSessionRepository(db)
	workspaceRepository := repository.NewWorkspaceRepository(db)
	workspaceMemberRepository := repository.NewWorkspaceMemberRepository(db)
	projectRepository := repository.NewProjectRepository(db)
	repositoryRepository := repository.NewRepositoryRepository(db)
	githubConnectionRepository := repository.NewGitHubConnectionRepository(db)
	githubImportJobRepository := repository.NewGitHubRepositoryImportJobRepository(db)

	// User
	userService := service.NewUserService(userRepository)
	userHandler := devflowhttp.NewUserHandler(userService)

	// Registration
	registrationService := service.NewRegistrationService(
		userRepository,
	)
	registrationHandler := devflowhttp.NewRegistrationHandler(
		registrationService,
	)

	// Authentication
	authService := service.NewAuthService(
		sessionRepository,
	)

	authMiddleware := devflowhttp.NewAuthMiddleware(
		authService,
	)

	// Session
	sessionService := service.NewSessionService(
		sessionRepository,
	)

	// Login
	loginService := service.NewLoginService(
		userRepository,
		sessionService,
	)
	loginHandler := devflowhttp.NewLoginHandler(
		loginService,
		cfg.CookieSecure,
	)

	// Workspace
	workspaceService := service.NewWorkspaceServiceWithMembers(
		workspaceRepository,
		workspaceMemberRepository,
	)
	workspaceHandler := devflowhttp.NewWorkspaceHandler(
		workspaceService,
	)

	// Workspace members
	workspaceMemberService := service.NewWorkspaceMemberService(
		workspaceMemberRepository,
		workspaceRepository,
	)
	workspaceMemberHandler := devflowhttp.NewWorkspaceMemberHandler(
		workspaceMemberService,
	)

	// Projects
	projectService := service.NewProjectService(
		projectRepository,
		workspaceMemberRepository,
	)
	projectHandler := devflowhttp.NewProjectHandler(
		projectService,
	)

	// Repositories
	repositoryService := service.NewRepositoryService(
		repositoryRepository,
		projectRepository,
		workspaceMemberRepository,
	)
	repositoryHandler := devflowhttp.NewRepositoryHandler(
		repositoryService,
	)

	// GitHub connections
	githubConnectionService := service.NewGitHubConnectionService(
		githubConnectionRepository,
		workspaceRepository,
		workspaceMemberRepository,
	)
	githubConnectionHandler := devflowhttp.NewGitHubConnectionHandler(
		githubConnectionService,
	)

	githubClient, githubClientErr := githubintegration.NewClient(
		cfg.GitHubApp,
		nil,
		"",
	)
	if githubClientErr != nil {
		log.Printf("GitHub integration unavailable: %v", githubClientErr)
		githubClient = githubintegration.NewUnavailableClient(githubClientErr)
	}

	githubRepositoryImportService := service.NewGitHubRepositoryImportService(
		projectRepository,
		githubConnectionService,
		githubClient,
		repositoryService,
	)
	githubRepositoryImportWorker := service.NewGitHubRepositoryImportWorkerWithStore(
		githubRepositoryImportService,
		32,
		service.GitHubRepositoryImportRetryPolicy{
			MaxAttempts: 3,
			Delay:       2 * time.Second,
		},
		githubImportJobRepository,
	)
	githubRepositoryImportHandler := devflowhttp.NewGitHubRepositoryImportHandler(
		githubRepositoryImportWorker,
	)
	jobContext, cancelJobs := context.WithCancel(ctx)
	defer cancelJobs()
	go githubRepositoryImportWorker.Start(jobContext)

	// Router
	router := devflowhttp.NewRouter(
		userHandler,
		registrationHandler,
		loginHandler,
		workspaceHandler,
		workspaceMemberHandler,
		projectHandler,
		repositoryHandler,
		githubRepositoryImportHandler,
		githubConnectionHandler,
		authMiddleware,
	)

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
