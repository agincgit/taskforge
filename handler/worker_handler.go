package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/agincgit/taskforge"
	"github.com/agincgit/taskforge/model"
)

type WorkerHandler struct {
	Manager *taskforge.Manager
}

func (h *WorkerHandler) RegisterWorker(c *gin.Context) {
	var ctx context.Context = c.Request.Context()
	var reg model.WorkerRegistration
	if err := c.ShouldBindJSON(&reg); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.Manager.RegisterWorker(ctx, &reg); err != nil {
		c.String(http.StatusInternalServerError, "Failed to register worker")
		return
	}
	c.JSON(http.StatusCreated, reg)
}

func (h *WorkerHandler) Heartbeat(c *gin.Context) {
	var ctx context.Context = c.Request.Context()
	id := c.Param("id")
	uuidVal, err := uuid.Parse(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := h.Manager.Heartbeat(ctx, uuidVal); err != nil {
		c.String(http.StatusNotFound, "Worker not found")
		return
	}
	c.Status(http.StatusOK)
}
