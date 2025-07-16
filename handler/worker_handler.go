package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
)

type WorkerHandler struct {
	DB *gorm.DB
}

func (h *WorkerHandler) RegisterWorker(c *gin.Context) {
	var reg model.WorkerRegistration
	if err := c.ShouldBindJSON(&reg); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.DB.Create(&reg).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to register worker")
		return
	}
	c.JSON(http.StatusCreated, reg)
}

func (h *WorkerHandler) Heartbeat(c *gin.Context) {
	id := c.Param("id")
	var beat model.WorkerHeartbeat
	if err := h.DB.First(&beat, "worker_id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Worker not found")
		return
	}
	beat.LastPing = time.Now()
	h.DB.Save(&beat)
	c.Status(http.StatusOK)
}
