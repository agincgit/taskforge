package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/agincgit/taskforge/pkg/model"
	"github.com/agincgit/taskforge/pkg/taskforge"
)

type TaskHandler struct {
	Manager *taskforge.Manager
}

func NewTaskHandler(mgr *taskforge.Manager) *TaskHandler {
	return &TaskHandler{Manager: mgr}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	ctx := c.Request.Context()
	var t model.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.String(http.StatusBadRequest, "Invalid body")
		return
	}
	if err := h.Manager.CreateTask(ctx, &t); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	ctx := c.Request.Context()
	tasks, err := h.Manager.GetTasks(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	uuidVal, err := uuid.Parse(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	t, err := h.Manager.GetTask(ctx, uuidVal)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	uuidVal, err := uuid.Parse(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	t, err := h.Manager.GetTask(ctx, uuidVal)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	if err := c.ShouldBindJSON(t); err != nil {
		c.String(http.StatusBadRequest, "Invalid body")
		return
	}
	if err := h.Manager.UpdateTask(ctx, t); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	uuidVal, err := uuid.Parse(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := h.Manager.DeleteTask(ctx, uuidVal); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
