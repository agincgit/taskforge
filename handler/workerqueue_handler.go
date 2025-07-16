package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
)

// WorkerQueueHandler manages worker queue operations.
type WorkerQueueHandler struct {
	DB *gorm.DB
}

// NewWorkerQueueHandler constructs a WorkerQueueHandler.
func NewWorkerQueueHandler(db *gorm.DB) *WorkerQueueHandler {
	return &WorkerQueueHandler{DB: db}
}

// EnqueueTask adds a new job to the queue.
func (h *WorkerQueueHandler) EnqueueTask(c *gin.Context) {
	var jq model.JobQueue
	if err := c.ShouldBindJSON(&jq); err != nil {
		c.String(http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.DB.Create(&jq).Error; err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, jq)
}

// GetQueue returns all queued jobs.
func (h *WorkerQueueHandler) GetQueue(c *gin.Context) {
	var queue []model.JobQueue
	if err := h.DB.Find(&queue).Error; err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, queue)
}

// DequeueTask removes a job from the queue by its ID.
func (h *WorkerQueueHandler) DequeueTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid ID parameter")
		return
	}
	if err := h.DB.Delete(&model.JobQueue{}, id).Error; err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
