package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
)

type TaskTemplateHandler struct {
	DB *gorm.DB
}

func NewTaskTemplateHandler(db *gorm.DB) *TaskTemplateHandler {
	return &TaskTemplateHandler{DB: db}
}

func (h *TaskTemplateHandler) CreateTaskTemplate(c *gin.Context) {
	var tmpl model.TaskTemplate
	if err := c.ShouldBindJSON(&tmpl); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.DB.Create(&tmpl).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to create template")
		return
	}
	c.JSON(http.StatusCreated, tmpl)
}

func (h *TaskTemplateHandler) GetTaskTemplates(c *gin.Context) {
	var tpls []model.TaskTemplate
	h.DB.Find(&tpls)
	c.JSON(http.StatusOK, tpls)
}

func (h *TaskTemplateHandler) UpdateTaskTemplate(c *gin.Context) {
	id := c.Param("id")
	var tmpl model.TaskTemplate
	if err := h.DB.First(&tmpl, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Template not found")
		return
	}
	if err := c.ShouldBindJSON(&tmpl); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}
	h.DB.Save(&tmpl)
	c.JSON(http.StatusOK, tmpl)
}

func (h *TaskTemplateHandler) DeleteTaskTemplate(c *gin.Context) {
	id := c.Param("id")
	h.DB.Delete(&model.TaskTemplate{}, "id = ?", id)
	c.Status(http.StatusNoContent)
}
