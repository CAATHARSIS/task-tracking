package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/CAATHARSIS/task-tracking/internal/models"
	board_repo "github.com/CAATHARSIS/task-tracking/internal/repository/board"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BoardHandler struct {
	repo      *board_repo.BoardPostgresRepo
	validator *validator.Validate
}

func NewBoardHandler(repo *board_repo.BoardPostgresRepo) *BoardHandler {
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
