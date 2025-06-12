package router

import (
	"net/http"

	"github.com/CAATHARSIS/task-tracking/internal/auth"
	"github.com/CAATHARSIS/task-tracking/internal/handlers/api"
	"github.com/CAATHARSIS/task-tracking/internal/handlers/web"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	apiBoardHandler *api.BoardHandler,
	webBoardHandler *web.BoardHandler,
	apiBoardTaskHandler *api.BoardTaskRelationHandler,
	apiTaskHandler *api.TaskHandler,
	webTaskHandler *web.TaskHandler,
	apiUserHandler *api.UserHandler,
	webUserHandler *web.UserHandler,
	jwtService *auth.JWTService,
) *gin.Engine {
	r := gin.Default()

	r.LoadHTMLFiles(
		"templates/home.html",
		"templates/base.html",
		"templates/error.html",
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
			authAPI.POST("/register", apiUserHandler.Register)
			authAPI.POST("/login", apiUserHandler.Login)
		}

		apiProtected := api.Group("")
		apiProtected.Use(jwtService.JWTAuthMiddleware())
		{
			userAPI := apiProtected.Group("/users")
			{
				userAPI.GET("/:id", apiUserHandler.GetUser)
				userAPI.PUT("/:id", apiUserHandler.UpdateUser)
				userAPI.DELETE("/:id", apiUserHandler.DeleteUser)
			}

			boardAPI := apiProtected.Group("/boards")
			{
				boardAPI.POST("", apiBoardHandler.CreateBoard)
				boardAPI.GET("/:id", apiBoardHandler.GetBoard)
				boardAPI.PUT("/:id", apiBoardHandler.UpdateBoard)
				boardAPI.DELETE("/:id", apiBoardHandler.DeleteBoard)
				boardAPI.GET("/:id/user-tasks", apiBoardHandler.ListBoardByUser)

				boardAPI.POST("/:id/tasks/:task_id", apiBoardTaskHandler.AddTaskToBoard)
				boardAPI.DELETE("/:id/tasks/:task_id", apiBoardTaskHandler.RemoveTaskFromBoard)
				boardAPI.GET("/:id/tasks", apiBoardTaskHandler.GetBoardTasks)
				boardAPI.PATCH("/tasks/move", apiBoardTaskHandler.MoveTasksBeetwenBoards)
			}

			taskAPI := apiProtected.Group("/tasks")
			{
				taskAPI.POST("", apiTaskHandler.CreateTask)
				taskAPI.GET("/:id", apiTaskHandler.GetTask)
				taskAPI.PATCH("/:id/status", apiTaskHandler.UpdateStatus)
				taskAPI.PUT("/:id", apiTaskHandler.UpdateTask)
				taskAPI.DELETE("/:id", apiTaskHandler.DeleteTask)
				taskAPI.GET("/user/:user_id", apiTaskHandler.ListTaskByUser)
			}
		}
	}

	web := r.Group("")
	{
		web.POST("/login", webUserHandler.LoginWeb)
		web.POST("/register", webUserHandler.RegisterWeb)

		webProtected := web.Group("")
		webProtected.Use(jwtService.JWTAuthMiddleware())
		{
			boardGroup := webProtected.Group("/boards")
			{
				boardGroup.GET("", webBoardHandler.ListBoardsPage)
				boardGroup.GET("/new", webBoardHandler.HandleBoardForm)
				boardGroup.GET("/:id", webBoardHandler.GetBoardPage)
				boardGroup.GET("/:id/edit", webBoardHandler.HandleBoardForm)
				boardGroup.POST("", webBoardHandler.HandleBoardForm)
				boardGroup.POST("/:id", webBoardHandler.HandleBoardForm)
				boardGroup.POST("/:id/delete", webBoardHandler.DeleteBoardWeb)

				boardGroup.POST("/:id/add-task", webBoardHandler.AddTaskToBoard)
				boardGroup.POST("/:id/create-and-add-task", webBoardHandler.CreateAndAddTaskToBoard)
				boardGroup.POST("/:id/remove-task/:task_id", webBoardHandler.RemoveTaskFromBoard)
			}

			taskGroup := webProtected.Group("/tasks")
			{
				taskGroup.GET("", webTaskHandler.ListTasksPage)
				taskGroup.GET("/new", webTaskHandler.HandleTaskForm)
				taskGroup.GET("/:id", webTaskHandler.GetTaskPage)
				taskGroup.GET("/:id/edit", webTaskHandler.HandleTaskForm)
				taskGroup.POST("", webTaskHandler.HandleTaskForm)
				taskGroup.POST("/:id", webTaskHandler.HandleTaskForm)
				taskGroup.POST("/:id/delete", webTaskHandler.DeleteTaskWeb)
				taskGroup.PATCH("/:id/status", apiTaskHandler.UpdateStatus)
			}
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	return r
}
