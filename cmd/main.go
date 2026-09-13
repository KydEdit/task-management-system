package main

import (
	"context"
	"log"
	"net/http"
	"task-manager-api/internal/config"
	"task-manager-api/internal/handler"
	"task-manager-api/internal/repository"
	"task-manager-api/internal/service"

	"github.com/jackc/pgx/v5"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	if err = conn.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	authService := service.NewAuthService(cfg.JWTSecret)
	// jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	userRepo := repository.NewUserRepository(conn)
	companyRepo := repository.NewCompanyRepository(conn)
	userService := service.NewUserService(userRepo, companyRepo)
	userHandler := handler.NewUserHandler(
		userService,
		authService,
	)

	taskRepo := repository.NewTaskRepository(conn)
	taskService := service.NewTaskUserService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	http.HandleFunc("/health", handler.HealthHandler)
	http.HandleFunc(
		"POST /users/register",
		userHandler.RegisterUser,
	)
	http.HandleFunc(
		"POST /users/login",
		userHandler.LoginUser,
	)
	http.HandleFunc(
		"GET /users/me",
		handler.AuthMiddleware(
			authService,
			userHandler.GetMe,
		),
	)

	http.HandleFunc(
		"POST /tasks",
		handler.AuthMiddleware(
			authService,
			taskHandler.CreateTask,
		),
	)

	http.HandleFunc(
		"GET /tasks",
		handler.AuthMiddleware(
			authService,
			taskHandler.GetTask,
		),
	)

	http.HandleFunc(
		"PUT /tasks/{id}",
		handler.AuthMiddleware(
			authService,
			taskHandler.EditTask,
		),
	)

	http.HandleFunc(
		"DELETE /tasks/{id}",
		handler.AuthMiddleware(
			authService,
			taskHandler.DeleteTask,
		),
	)

	log.Println("Starting server on: 8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
