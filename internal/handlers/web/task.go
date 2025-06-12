package web

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
	isNew := idStr == "" || idStr == "new"

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

	if isNew {
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

		if task.UserID != userID {
			c.HTML(http.StatusForbidden, "error.html", gin.H{"error": "Access denied"})
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

func (h *TaskHandler) DeleteTaskWeb(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"TemplateName": "erorr",
			"error":        "Invalid task ID",
		})
		return
	}

	userID := c.MustGet("user_id").(int)
	task, err := h.repo.GetById(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Task not found",
		})
		return
	}

	if task.UserID != userID {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error": "Access denied",
		})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed with deleting",
		})
		return
	}

	c.Redirect(http.StatusFound, "/tasks")
}
