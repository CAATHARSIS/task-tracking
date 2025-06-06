package task

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CAATHARSIS/task-tracking/internal/models"
	task_repo "github.com/CAATHARSIS/task-tracking/internal/repository/task"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TaskHandler struct {
	repo      *task_repo.TaskPostgresRepo
	validator *validator.Validate
}

func NewTaskHandler(repo *task_repo.TaskPostgresRepo) *TaskHandler {
	v := validator.New()
	v.RegisterValidation("taskstatus", func(fl validator.FieldLevel) bool {
		status := fl.Field().Interface().(models.TaskStatus)
		return status.IsValid()
	})

	return &TaskHandler{
		repo:      repo,
		validator: v,
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req models.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": c.Keys})
		return
	}
	req.UserID = userID.(int)

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newTask := models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		UserID:      req.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.Create(&newTask); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newTask)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.repo.GetById(id)
	if err != nil {
		if err.Error() == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var update models.TaskStatusUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.repo.GetById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	task.Status = update.Status
	task.UpdatedAt = time.Now()

	if err := h.repo.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	task.ID = id
	task.UpdatedAt = time.Now()

	if err := h.validator.Struct(task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Update(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedTask, err := h.repo.GetById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedTask)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TaskHandler) ListTaskByUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	tasks, err := h.repo.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) ListTasksPage(c *gin.Context) {
	userID := c.MustGet("user_id")

	tasks, err := h.repo.ListByUser(userID.(int))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"TemplateName": "tasks-list",
			"error":        err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "tasks-list.html", gin.H{
		"TemplateName":    "tasks-list",
		"Tasks":           tasks,
		"IsAuthenticated": true,
	})
}

func (h *TaskHandler) GetTaskPage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"TemplateName": "tasks-view",
			"error":        "Invalid task ID",
		})
		return
	}

	task, err := h.repo.GetById(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"TemplateName": "tasks-view",
			"error":        "Failed to load task",
		})
		return
	}

	userID, _ := c.Get("user_id")
	if task.UserID != userID.(int) {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"TemplateName": "tasks-view",
			"error":        "Access is unavailable",
		})
		return
	}

	c.HTML(http.StatusOK, "tasks-view.html", gin.H{
		"TemplateName":    "tasks-view",
		"Task":            task,
		"IsAuthenticated": true,
	})
}

func (h *TaskHandler) HandleTaskForm(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		idStr := c.Param("id")
		if idStr == "" || idStr == "new" {
			c.HTML(http.StatusOK, "tasks-form.html", gin.H{
				"TemplateName": "tasks-form",
				"IsNew":        true,
			})
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid task ID"})
			return
		}

		task, err := h.repo.GetById(id)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Task not found"})
			return
		}

		c.HTML(http.StatusOK, "tasks-form.html", gin.H{
			"TemplateName": "tasks-form",
			"Task":         task,
			"IsNew":        false,
		})
		return
	}

	idStr := c.Param("id")
	userID := c.MustGet("user_id").(int)

	title := strings.TrimSpace(c.PostForm("title"))
	description := strings.TrimSpace(c.PostForm("description"))
	status := c.PostForm("status")


	if title == "" || len(title) > 100 {
		c.HTML(http.StatusBadRequest, "tasks-form.html", gin.H{
			"error": "Название задачи обязательно (макс. 100 символов)",
			"Task":  &models.Task{Title: title, Description: description, Status: models.TaskStatus(status)},
			"IsNew": idStr == "" || idStr == "new",
		})
		return
	}

	if len(description) > 500 {
		c.HTML(http.StatusBadRequest, "tasks-form.html", gin.H{
			"error": "Описание не должно превышать 500 символов",
			"Task":  &models.Task{Title: title, Description: description, Status: models.TaskStatus(status)},
			"IsNew": idStr == "" || idStr == "new",
		})
		return
	}

	if !models.TaskStatus(status).IsValid() {
		c.HTML(http.StatusBadRequest, "tasks-form.html", gin.H{
			"error": "Недопустимый статус задачи",
			"Task":  &models.Task{Title: title, Description: description, Status: models.TaskStatus(status)},
			"IsNew": idStr == "" || idStr == "new",
		})
		return
	}

	if idStr == "" || idStr == "new" {
		newTask := models.Task{
			Title:       title,
			Description: description,
			Status:      models.TaskStatus(status),
			UserID:      userID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := h.repo.Create(&newTask); err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}
	} else {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
			return
		}

		task, err := h.repo.GetById(id)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Task not found"})
			return
		}

		task.Title = title
		task.Description = description
		task.Status = models.TaskStatus(status)
		task.UpdatedAt = time.Now()

		if err := h.repo.Update(task); err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}
	}

	c.Redirect(http.StatusFound, "/tasks")
}
