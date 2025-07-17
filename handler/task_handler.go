package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/agincgit/taskforge"
	"github.com/agincgit/taskforge/model"
)

type TaskHandler struct {
	Manager *taskforge.Manager
}

func NewTaskHandler(mgr *taskforge.Manager) *TaskHandler {
	return &TaskHandler{Manager: mgr}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var t model.Task
	if err := c.ShouldBindJSON(t); err != nil {
		c.String(http.StatusBadRequest, "Invalid body")
		return
	}
	if err := h.Manager.CreateTask(c.Request.Context(), &t); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	tasks, err := h.Manager.GetTasks(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	uuidVal, err := uuid.Parse(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	t, err := h.Manager.GetTask(c.Request.Context(), uuidVal)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	uuidVal, err := uuid.Parse(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	t, err := h.Manager.GetTask(c.Request.Context(), uuidVal)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	if err := c.ShouldBindJSON(&t); err != nil {
		c.String(http.StatusBadRequest, "Invalid body")
		return
	}
	if err := h.Manager.UpdateTask(c.Request.Context(), t); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	uuidVal, err := uuid.Parse(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := h.Manager.DeleteTask(c.Request.Context(), uuidVal); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
