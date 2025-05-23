package board

import (
	"net/http"
	"strconv"

	board_task_repo "github.com/CAATHARSIS/task-tracking/internal/repository/board_task"
	"github.com/gin-gonic/gin"
)

type BoardTaskRelationHandler struct {
	repo *board_task_repo.BoardTaskPostgresRepo
}

func NewBoardTaskRealtionHandler(repo *board_task_repo.BoardTaskPostgresRepo) *BoardTaskRelationHandler {
	return &BoardTaskRelationHandler{repo: repo}
}

func (h *BoardTaskRelationHandler) AddTaskToBoard(c *gin.Context) {
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	taskID, err := strconv.Atoi(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.repo.AddTask(boardID, taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BoardTaskRelationHandler) RemoveTaskFromBoard(c *gin.Context) {
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	taskID, err := strconv.Atoi(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.repo.RemoveTask(boardID, taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BoardTaskRelationHandler) GetBoardTasks(c *gin.Context) {
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	taskIDs, err := h.repo.GetTasks(boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, taskIDs)
}

func (h *BoardTaskRelationHandler) MoveTasksBeetwenBoards(c *gin.Context) {
	var req struct {
		FromBoardID int `json:"from_board_id" binding:"required"`
		ToBoardID   int `json:"to_board_id" binding:"required"`
		TaskID      int `json:"task_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.MoveTask(req.FromBoardID, req.ToBoardID, req.TaskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
