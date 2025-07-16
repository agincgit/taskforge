package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
)

type TaskHandler struct {
	DB *gorm.DB
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{DB: db}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var t model.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.String(http.StatusBadRequest, "Invalid body")
		return
	}
	h.DB.Create(&t)
	c.JSON(http.StatusCreated, t)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	var tasks []model.Task
	h.DB.Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	var t model.Task
	if err := h.DB.First(&t, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var t model.Task
	if h.DB.First(&t, "id = ?", id).Error != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	if err := c.ShouldBindJSON(&t); err != nil {
		c.String(http.StatusBadRequest, "Invalid body")
		return
	}
	h.DB.Save(&t)
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	h.DB.Delete(&model.Task{}, "id = ?", id)
	c.Status(http.StatusNoContent)
}
