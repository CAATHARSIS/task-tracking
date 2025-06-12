package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CAATHARSIS/task-tracking/internal/models"
	board_repo "github.com/CAATHARSIS/task-tracking/internal/repository/board"
	board_task_repo "github.com/CAATHARSIS/task-tracking/internal/repository/board_task"
	task_repo "github.com/CAATHARSIS/task-tracking/internal/repository/task"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BoardHandler struct {
	repo          *board_repo.BoardPostgresRepo
	boardTaskRepo *board_task_repo.BoardTaskPostgresRepo
	taskRepo      *task_repo.TaskPostgresRepo
	validator     *validator.Validate
}

func NewBoardHandler(repo *board_repo.BoardPostgresRepo,
	boardTaskRepo *board_task_repo.BoardTaskPostgresRepo,
	taskRepo *task_repo.TaskPostgresRepo) *BoardHandler {
	v := validator.New()
	return &BoardHandler{
		repo:          repo,
		boardTaskRepo: boardTaskRepo,
		taskRepo:      taskRepo,
		validator:     v,
	}
}

func (h *BoardHandler) ListBoardsPage(c *gin.Context) {
	userID := c.MustGet("user_id")

	boards, err := h.repo.ListByUser(userID.(int))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"TemplateName": "boards-list",
			"error":        err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "boards-list.html", gin.H{
		"TemplateName":    "boards-list",
		"Boards":          boards,
		"IsAuthenticated": true,
	})
}

func (h *BoardHandler) HandleBoardForm(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		idStr := c.Param("id")
		if idStr == "" || idStr == "new" {
			c.HTML(http.StatusOK, "boards-form.html", gin.H{
				"TemplateName": "boards-form",
				"IsNew":        true,
			})
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid board ID"})
			return
		}

		task, err := h.repo.GetById(id)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Board not found"})
			return
		}

		c.HTML(http.StatusOK, "boards-form.html", gin.H{
			"TemplateName": "boards-form",
			"Board":        task,
			"IsNew":        false,
		})
		return
	}

	idStr := c.Param("id")
	userID := c.MustGet("user_id").(int)
	isNew := idStr == "" || idStr == "new"

	name := strings.TrimSpace(c.PostForm("name"))

	if name == "" || len(name) > 50 {
		c.HTML(http.StatusBadRequest, "boards-form.html", gin.H{
			"error": "Название доски обязательно (макс. 50 символов)",
			"Task":  &models.Board{Name: name},
			"IsNew": idStr == "" || idStr == "new",
		})
		return
	}

	if isNew {
		newBoard := models.Board{
			Name:      name,
			UserID:    userID,
			CreatedAt: time.Now(),
		}

		if err := h.repo.Create(&newBoard); err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}
	} else {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
			return
		}

		board, err := h.repo.GetById(id)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Task not found"})
			return
		}

		if board.UserID != userID {
			c.HTML(http.StatusForbidden, "error.html", gin.H{"error": "Access denied"})
			return
		}

		board.Name = name

		if err := h.repo.Update(board); err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}
	}

	c.Redirect(http.StatusFound, "/boards")
}

func (h *BoardHandler) DeleteBoardWeb(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"TemplateName": "erorr",
			"error":        "Invalid task ID",
		})
		return
	}

	userID := c.MustGet("user_id").(int)
	board, err := h.repo.GetById(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Task not found",
		})
		return
	}

	if board.UserID != userID {
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

	c.Redirect(http.StatusFound, "/boards")
}

func (h *BoardHandler) GetBoardPage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"TemplateName": "boards-view",
			"error":        err.Error(),
		})
		return
	}

	board, err := h.repo.GetById(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"TemplateName": "boards-view",
			"error":        err.Error(),
		})
	}

	userID := c.MustGet("user_id")
	if board.UserID != userID.(int) {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"TemplateName": "boards-view",
			"error":        "Access denied",
		})
		return
	}

	tasksIDs, err := h.boardTaskRepo.GetTasks(id)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"TemplateName": "boards-view",
			"error":        "Failed to load task",
		})
	}

	var tasks []models.Task
	for _, taskID := range tasksIDs {
		task, err := h.taskRepo.GetById(taskID)
		if err != nil {
			continue
		}
		tasks = append(tasks, *task)
	}

	userTasks, err := h.taskRepo.ListByUser(userID.(int))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"TemplateName": "boards-view",
			"error":        "Failed to load user tasks",
		})
		return
	}

	c.HTML(http.StatusOK, "boards-view.html", gin.H{
		"TemplateName":    "boards-view",
		"Board":           board,
		"Tasks":           tasks,
		"UserTasks":       userTasks,
		"IsAuthenticated": true,
	})
}

func (h *BoardHandler) AddTaskToBoard(c *gin.Context) {
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid board ID"})
		return
	}

	taskID, err := strconv.Atoi(c.PostForm("task_id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid task ID"})
		return
	}

	userID := c.MustGet("user_id").(int)

	board, err := h.repo.GetById(boardID)
	if err != nil || board.UserID != userID {
		c.HTML(http.StatusForbidden, "error.html", gin.H{"error": "Access denied"})
		return
	}

	task, err := h.taskRepo.GetById(taskID)
	if err != nil || task.UserID != userID {
		c.HTML(http.StatusForbidden, "error.html", gin.H{"error": "Access denied"})
		return
	}

	if err := h.boardTaskRepo.AddTask(boardID, taskID); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/boards/"+strconv.Itoa(boardID))
}

func (h *BoardHandler) CreateAndAddTaskToBoard(c *gin.Context) {
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid board ID"})
		return
	}

	userID := c.MustGet("user_id").(int)

	board, err := h.repo.GetById(boardID)
	if err != nil || board.UserID != userID {
		c.HTML(http.StatusForbidden, "error.html", gin.H{"error": "Access denied"})
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	description := strings.TrimSpace(c.PostForm("description"))
	status := c.PostForm("status")

	if title == "" || len(title) > 100 {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Название задачи обязательно (макс. 100 символов)"})
		return
	}

	if len(description) > 500 {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Описание не должно превышать 500 символов"})
		return
	}

	if !models.TaskStatus(status).IsValid() {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Недопустимый статус задачи"})
		return
	}

	newTask := models.Task{
		Title:       title,
		Description: description,
		Status:      models.TaskStatus(status),
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.taskRepo.Create(&newTask); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	if err := h.boardTaskRepo.AddTask(boardID, newTask.ID); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/boards/"+strconv.Itoa(boardID))
}

func (h *BoardHandler) RemoveTaskFromBoard(c *gin.Context) {
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid board ID"})
		return
	}

	taskID, err := strconv.Atoi(c.Param("task_id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid task ID"})
		return
	}

	userID := c.MustGet("user_id").(int)

	board, err := h.repo.GetById(boardID)
	if err != nil || board.UserID != userID {
		c.HTML(http.StatusForbidden, "error.html", gin.H{"error": "Access denied"})
		return
	}

	if err := h.boardTaskRepo.RemoveTask(boardID, taskID); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/boards/"+strconv.Itoa(boardID))
}
