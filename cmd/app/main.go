package main

import (
	"log"

	"github.com/CAATHARSIS/task-tracking/internal/auth"
	"github.com/CAATHARSIS/task-tracking/internal/config"
	"github.com/CAATHARSIS/task-tracking/internal/handlers/api"
	"github.com/CAATHARSIS/task-tracking/internal/handlers/web"
	board_repo "github.com/CAATHARSIS/task-tracking/internal/repository/board"
	board_task_repo "github.com/CAATHARSIS/task-tracking/internal/repository/board_task"
	task_repo "github.com/CAATHARSIS/task-tracking/internal/repository/task"
	user_repo "github.com/CAATHARSIS/task-tracking/internal/repository/user"
	"github.com/CAATHARSIS/task-tracking/internal/router"
	"github.com/CAATHARSIS/task-tracking/pkg/database"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("PostgreSQL connection error: %v", err)
	}
	defer db.Close()

	jwtService := auth.NewJWTService(cfg)

	boardRepo := board_repo.NewBoardPostgresRepo(db)
	boardTaskRepo := board_task_repo.NewBoardTaskPostgresRepo(db)
	taskRepo := task_repo.NewTaskPostgresRepo(db)
	userRepo := user_repo.NewUserPostgresRepo(db)

	apiBoardHandler := api.NewBoardHandler(boardRepo)
	webBoardHandler := web.NewBoardHandler(boardRepo, boardTaskRepo, taskRepo)
	apiBoardTaskHandler := api.NewBoardTaskRealtionHandler(boardTaskRepo)
	apiTaskHandler := api.NewTaskHandler(taskRepo)
	webTaskHandler := web.NewTaskHandler(taskRepo)
	apiUserHandler := api.NewUserHandler(userRepo)
	webUserHandler := web.NewUserHandler(userRepo, taskRepo)

	r := router.SetupRouter(
		apiBoardHandler,
		webBoardHandler,
		apiBoardTaskHandler,
		apiTaskHandler,
		webTaskHandler,
		apiUserHandler,
		webUserHandler,
		jwtService,
	)

	r.Run(":" + cfg.AppPort)
}
