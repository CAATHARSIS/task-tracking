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

	r.LoadHTMLFiles(
		"templates/home.html",
		"templates/base.html",
		"templates/auth/login.html",
		"templates/auth/register.html",
		"templates/boards/boards-form.html",
		"templates/boards/boards-list.html",
		"templates/boards/boards-view.html",
		"templates/tasks/tasks-form.html",
		"templates/tasks/tasks-list.html",
		"templates/tasks/tasks-view.html",
	)
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"TemplateName": "home",
		})
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"TemplateName": "login",
		})
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"TemplateName": "register",
		})
	})

	api := r.Group("/api")
	{
		authAPI := api.Group("/auth")
		{
			authAPI.POST("/register", userHandler.Register)
			authAPI.POST("/login", userHandler.Login)
		}

		apiProtected := api.Group("")
		apiProtected.Use(jwtService.JWTAuthMiddleware())
		{
			userAPI := apiProtected.Group("/users")
			{
				userAPI.GET("/:id", userHandler.GetUser)
				userAPI.PUT("/:id", userHandler.UpdateUser)
				userAPI.DELETE("/:id", userHandler.DeleteUser)
			}

			boardAPI := apiProtected.Group("/boards")
			{
				boardAPI.POST("", boardHandler.CreateBoard)
				boardAPI.GET("/:id", boardHandler.GetBoard)
				boardAPI.PUT("/:id", boardHandler.UpdateBoard)
				boardAPI.DELETE("/:id", boardHandler.DeleteBoard)
				boardAPI.GET("/:id/user-tasks", boardHandler.ListBoardByUser)

				boardAPI.POST("/:id/tasks/:task_id", boardTaskHandler.AddTaskToBoard)
				boardAPI.DELETE("/:id/tasks/:task_id", boardTaskHandler.RemoveTaskFromBoard)
				boardAPI.GET("/:id/tasks", boardTaskHandler.GetBoardTasks)
				boardAPI.PATCH("/tasks/move", boardTaskHandler.MoveTasksBeetwenBoards)
			}

			taskAPI := apiProtected.Group("/tasks")
			{
				taskAPI.POST("", taskHandler.CreateTask)
				taskAPI.GET("/:id", taskHandler.GetTask)
				taskAPI.PATCH("/:id/status", taskHandler.UpdateStatus)
				taskAPI.PUT("/:id", taskHandler.UpdateTask)
				taskAPI.DELETE("/:id", taskHandler.DeleteTask)
				taskAPI.GET("/user/:user_id", taskHandler.ListTaskByUser)
			}
		}
	}

	web := r.Group("")
	{
		web.POST("/login", userHandler.LoginWeb)
		web.POST("/register", userHandler.RegisterWeb)

		webProtected := web.Group("")
		webProtected.Use(jwtService.JWTAuthMiddleware())
		{
			webProtected.GET("/boards", boardHandler.ListBoardsPage)
			webProtected.GET("/boards/new", boardHandler.HandleBoardForm)
			webProtected.GET("/boards/:id", boardHandler.GetBoardPage)
			webProtected.GET("/boards/:id/edit", boardHandler.HandleBoardForm)
			webProtected.POST("/boards", boardHandler.HandleBoardForm)
			webProtected.POST("/boards/:id", boardHandler.HandleBoardForm)

			webProtected.GET("/tasks", taskHandler.ListTasksPage)
			webProtected.GET("/tasks/new", taskHandler.HandleTaskForm)
			webProtected.GET("/tasks/:id", taskHandler.GetTaskPage)
			webProtected.GET("/tasks/:id/edit", taskHandler.HandleTaskForm)
			webProtected.POST("/tasks", taskHandler.HandleTaskForm)
			webProtected.POST("/tasks:id", taskHandler.HandleTaskForm)
			webProtected.PATCH("/tasks/:id/status", taskHandler.UpdateStatus)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	return r
}
