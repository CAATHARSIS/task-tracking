package user

import (
	"net/http"
	"strconv"

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

type BaseAuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest BaseAuthRequest
type LoginRequest BaseAuthRequest

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.repo.GetByEmail(req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := h.repo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "email": user.Email})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid email or password"})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to load config"})
		return
	}

	s := auth.NewJWTService(cfg)
	token, err := s.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.repo.GetById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=6"`
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"user": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.GetById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.PasswordHash = hashedPassword
	}

	if err := h.repo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated`"})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func (h *UserHandler) LoginWeb(c *gin.Context) {
	// Получаем данные из формы
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Валидация
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

	// Валидация
	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"TemplateName": "register",
			"error":        "Email и пароль обязательны",
		})
		return
	}

	// Проверка существования пользователя
	if _, err := h.repo.GetByEmail(email); err == nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"TemplateName": "register",
			"error":        "Пользователь с таким email уже существует",
		})
		return
	}

	// Создание пользователя
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
