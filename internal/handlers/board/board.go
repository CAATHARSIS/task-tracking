package board

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
		repo:      repo,
		validator: v,
	}
}

func (h *BoardHandler) CreateBoard(c *gin.Context) {
	var req models.BoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": c.Keys})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBoard := models.Board{
		Name:       req.Name,
		UserID:     userID.(int),
		CreatedAt:  time.Now(),
		UpdateddAt: time.Now(),
	}

	if err := h.repo.Create(&newBoard); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newBoard)
}

func (h *BoardHandler) GetBoard(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	board, err := h.repo.GetById(id)
	if err != nil {
		if err.Error() == "board not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, board)
}

func (h *BoardHandler) UpdateBoard(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	var req models.BoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	currentBoard, err := h.repo.GetById(id)
	if err != nil {
		if err.Error() == "board not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	currentBoard.Name = req.Name
	currentBoard.UpdateddAt = time.Now()

	if err := h.repo.Update(currentBoard); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, currentBoard)
}

func (h *BoardHandler) DeleteBoard(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BoardHandler) ListBoardByUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	boards, err := h.repo.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, boards)
}

func (h *BoardHandler) ListBoardsPage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

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

func (h *BoardHandler) GetBoardPage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"TemplateName": "boards-view",
			"error":        "Invalid board ID",
		})
		return
	}

	board, err := h.repo.GetById(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"TemplateName": "boards-view",
			"error":        err.Error(),
		})
		return
	}

	userID, _ := c.Get("user_id")
	if board.UserID != userID.(int) {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"TemplateName": "boards-view",
			"error":        "Access is unavailable",
		})
		return
	}

	tasksIDs, err := h.boardTaskRepo.GetTasks(id)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"TemplateName": "boards-view",
			"error":        "Failed to load task",
		})
		return
	}

	var tasks []models.Task
	for _, taskID := range tasksIDs {
		task, err := h.taskRepo.GetById(taskID)
		if err != nil {
			continue
		}
		tasks = append(tasks, *task)
	}

	c.HTML(http.StatusOK, "boards-view.html", gin.H{
		"TemplateName":    "boards-view",
		"Board":           board,
		"Tasks":           tasks,
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

		board, err := h.repo.GetById(id)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Board not found"})
			return
		}

		c.HTML(http.StatusOK, "boards-form.html", gin.H{
			"TemplateName": "boards-form",
			"Board":        board,
			"IsNew":        false,
		})
		return
	}

	idStr := c.Param("id")
	userID := c.MustGet("user_id").(int)
	name := strings.TrimSpace(c.PostForm("name"))

	if name == "" || len(name) < 3 || len(name) > 50 {
        c.HTML(http.StatusBadRequest, "boards-form.html", gin.H{
            "error": "Название доски должно быть от 3 до 50 символов",
            "Board": &models.Board{Name: name},
            "IsNew": idStr == "" || idStr == "new",
        })
        return
    }

	if idStr == "" || idStr == "new" {
		newBoard := models.Board{
			Name:       name,
			UserID:     userID,
			CreatedAt:  time.Now(),
			UpdateddAt: time.Now(),
		}

		if err := h.repo.Create(&newBoard); err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}
	} else {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid board ID"})
			return
		}

		board, err := h.repo.GetById(id)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Board not found"})
			return
		}

		board.Name = name
		board.UpdateddAt = time.Now()

		if err := h.repo.Update(board); err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}
	}

	c.Redirect(http.StatusFound, "/boards")
}
