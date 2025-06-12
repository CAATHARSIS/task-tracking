package web

import (
	"net/http"

	"github.com/CAATHARSIS/task-tracking/internal/auth"
	"github.com/CAATHARSIS/task-tracking/internal/config"
	"github.com/CAATHARSIS/task-tracking/internal/models"
	task_repo "github.com/CAATHARSIS/task-tracking/internal/repository/task"
	user_repo "github.com/CAATHARSIS/task-tracking/internal/repository/user"
	"github.com/CAATHARSIS/task-tracking/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	repo      *user_repo.UserPostgrtesRepo
	taskRepo  *task_repo.TaskPostgresRepo
	validator *validator.Validate
}

func NewUserHandler(repo *user_repo.UserPostgrtesRepo, taskRepo *task_repo.TaskPostgresRepo) *UserHandler {
	return &UserHandler{
		repo:      repo,
		taskRepo:  taskRepo,
		validator: validator.New(),
	}
}

func (h *UserHandler) LoginWeb(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"TemplateName": "login",
			"error":        "Email и пароль обязательны",
		})
		return
	}

	user, err := h.repo.GetByEmail(email)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"TemplateName": "login",
			"error":        "Invalid credentials",
		})
		return
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"TemplateName": "login",
			"error":        "Неверный email или пароль",
		})
		return
	}

	cfg, _ := config.Load()
	service := auth.NewJWTService(cfg)
	token, err := service.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.SetCookie("auth_token", token, 3600, "/", "", false, true)
	c.Redirect(http.StatusFound, "/tasks")
}

func (h *UserHandler) RegisterWeb(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"TemplateName": "register",
			"error":        "Email и пароль обязательны",
		})
		return
	}

	if _, err := h.repo.GetByEmail(email); err == nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"TemplateName": "register",
			"error":        "Пользователь с таким email уже существует",
		})
		return
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"TemplateName": "register",
			"error":        "Ошибка сервера",
		})
		return
	}
	user := &models.User{
		Email:        email,
		PasswordHash: hashedPassword,
	}

	if err := h.repo.Create(user); err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"TemplateName": "register",
			"error":        "Server error",
		})
		return
	}

	cfg, _ := config.Load()
	service := auth.NewJWTService(cfg)

	token, err := service.GenerateJWT(user.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"TemplateName": "register",
			"error":        "Authorization error",
		})
		return
	}

	c.SetCookie("auth_token", token, 3600*24*7, "/", "", false, true)

	tasks, _ := h.taskRepo.ListByUser(user.ID)
	c.HTML(http.StatusOK, "tasks-list.html", gin.H{
		"TemplateName":    "tasks-list",
		"Tasks":           tasks,
		"IsAuthenticated": true,
	})
}
