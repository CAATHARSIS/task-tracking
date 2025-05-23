package router

import (
	"net/http"

	"github.com/CAATHARSIS/task-tracking/internal/auth"
	"github.com/CAATHARSIS/task-tracking/internal/handlers/board"
	"github.com/CAATHARSIS/task-tracking/internal/handlers/task"
	"github.com/CAATHARSIS/task-tracking/internal/handlers/user"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	boardHandler *board.BoardHandler,
	boardTaskHandler *board.BoardTaskRelationHandler,
	taskHandler *task.TaskHandler,
	userHandler *user.UserHandler,
	jwtService *auth.JWTService,
) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", userHandler.Login)
		}

		protected := api.Group("")
		protected.Use(jwtService.JWTAuthMiddleware())
		{
			userGroup := protected.Group("/users")
			{
				userGroup.GET("/:id", userHandler.GetUser)
				userGroup.PUT("/:id", userHandler.UpdateUser)
				userGroup.DELETE("/:id", userHandler.DeleteUser)
			}

			boardGroup := protected.Group("/boards")
			{
				boardGroup.POST("", boardHandler.CreateBoard)
				boardGroup.GET("/:id", boardHandler.GetBoard)
				boardGroup.PUT("/:id", boardHandler.UpdateBoard)
				boardGroup.DELETE("/:id", boardHandler.DeleteBoard)
				boardGroup.GET("/:id/user-tasks", boardHandler.ListBoardByUser)

				boardGroup.POST("/:id/tasks/:task_id", boardTaskHandler.AddTaskToBoard)
				boardGroup.DELETE("/:id/tasks/:task_id", boardTaskHandler.RemoveTaskFromBoard)
				boardGroup.GET("/:id/tasks", boardTaskHandler.GetBoardTasks)
				boardGroup.PATCH("/tasks/move", boardTaskHandler.MoveTasksBeetwenBoards)
			}

			taskGroup := protected.Group("/tasks")
			{
				taskGroup.POST("", taskHandler.CreateTask)
				taskGroup.GET("/:id", taskHandler.GetTask)
				taskGroup.PATCH("/:id/status", taskHandler.UpdateStatus)
				taskGroup.PUT("/:id", taskHandler.UpdateTask)
				taskGroup.DELETE("/:id", taskHandler.DeleteTask)
				// taskGroup.GET("/:id/tasks", taskHandler.ListTaskByUser)
			}
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	return r
}
