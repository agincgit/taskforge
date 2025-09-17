package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/agincgit/taskforge"
	"github.com/agincgit/taskforge/model"
)

type TaskTemplateHandler struct {
	Manager   *taskforge.Manager
	Scheduler TemplateScheduler
}

type TemplateScheduler interface {
	OnTemplateChanged(tpl model.TaskTemplate) error
	OnTemplateDeleted(templateID uuid.UUID)
}

func NewTaskTemplateHandler(mgr *taskforge.Manager, sched TemplateScheduler) *TaskTemplateHandler {
	return &TaskTemplateHandler{Manager: mgr, Scheduler: sched}
}

func (h *TaskTemplateHandler) CreateTaskTemplate(c *gin.Context) {
	var ctx context.Context = c.Request.Context()
	var tmpl model.TaskTemplate
	if err := c.ShouldBindJSON(&tmpl); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.Manager.CreateTaskTemplate(ctx, &tmpl); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create template")
		return
	}
	if h.Scheduler != nil {
		if err := h.Scheduler.OnTemplateChanged(tmpl); err != nil {
			c.String(http.StatusInternalServerError, "Failed to register schedule")
			return
		}
	}
	c.JSON(http.StatusCreated, tmpl)
}

func (h *TaskTemplateHandler) GetTaskTemplates(c *gin.Context) {
	var ctx context.Context = c.Request.Context()
	tpls, err := h.Manager.GetTaskTemplates(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, tpls)
}

func (h *TaskTemplateHandler) UpdateTaskTemplate(c *gin.Context) {
	var ctx context.Context = c.Request.Context()
	id := c.Param("id")
	uuidVal, err := uuid.Parse(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	tmpl, err := h.Manager.GetTaskTemplate(ctx, uuidVal)
	if err != nil {
		c.String(http.StatusNotFound, "Template not found")
		return
	}
	if err := c.ShouldBindJSON(tmpl); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.Manager.UpdateTaskTemplate(ctx, tmpl); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if h.Scheduler != nil {
		if err := h.Scheduler.OnTemplateChanged(*tmpl); err != nil {
			c.String(http.StatusInternalServerError, "Failed to register schedule")
			return
		}
	}
	c.JSON(http.StatusOK, tmpl)
}

func (h *TaskTemplateHandler) DeleteTaskTemplate(c *gin.Context) {
	var ctx context.Context = c.Request.Context()
	id := c.Param("id")
	uuidVal, err := uuid.Parse(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := h.Manager.DeleteTaskTemplate(ctx, uuidVal); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if h.Scheduler != nil {
		h.Scheduler.OnTemplateDeleted(uuidVal)
	}
	c.Status(http.StatusNoContent)
}
