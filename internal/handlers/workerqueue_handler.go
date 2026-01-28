package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/agincgit/taskforge/pkg/model"
	"github.com/agincgit/taskforge/pkg/taskforge"
)

// WorkerQueueHandler manages worker queue operations.
type WorkerQueueHandler struct {
	Manager *taskforge.Manager
}

// NewWorkerQueueHandler constructs a WorkerQueueHandler.
func NewWorkerQueueHandler(mgr *taskforge.Manager) *WorkerQueueHandler {
	return &WorkerQueueHandler{Manager: mgr}
}

// EnqueueTask adds a new job to the queue.
func (h *WorkerQueueHandler) EnqueueTask(c *gin.Context) {
	ctx := c.Request.Context()
	var jq model.JobQueue
	if err := c.ShouldBindJSON(&jq); err != nil {
		c.String(http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.Manager.EnqueueJob(ctx, &jq); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, jq)
}

// GetQueue returns all queued jobs.
func (h *WorkerQueueHandler) GetQueue(c *gin.Context) {
	ctx := c.Request.Context()
	queue, err := h.Manager.GetQueue(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, queue)
}

// DequeueTask removes a job from the queue by its ID.
func (h *WorkerQueueHandler) DequeueTask(c *gin.Context) {
	ctx := c.Request.Context()
	idParam := c.Param("id")
	uuidVal, err := uuid.Parse(idParam)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid ID parameter")
		return
	}
	if err := h.Manager.DequeueJob(ctx, uuidVal); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
