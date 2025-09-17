package server

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	taskforge "github.com/agincgit/taskforge"
	"github.com/agincgit/taskforge/handler"
)

// NewRouter sets up database migrations and registers all TaskForge API routes.
func NewRouter(db *gorm.DB) (*gin.Engine, error) {
	mgr, err := taskforge.NewManager(taskforge.TaskForgeConfig{
		DB:        db,
		TableName: "tasks",
		Context:   context.Background(),
	})
	if err != nil {
		return nil, err
	}

	router := gin.Default()
	api := router.Group("/taskforge/api/v1")

	// Task endpoints
	th := handler.NewTaskHandler(mgr)
	api.POST("/tasks", th.CreateTask)
	api.GET("/tasks", th.GetTasks)
	api.GET("/tasks/:id", th.GetTask)
	api.PUT("/tasks/:id", th.UpdateTask)
	api.DELETE("/tasks/:id", th.DeleteTask)

	// WorkerQueue endpoints
	wqh := handler.NewWorkerQueueHandler(mgr)
	api.POST("/workerqueue", wqh.EnqueueTask)
	api.GET("/workerqueue", wqh.GetQueue)
	api.DELETE("/workerqueue/:id", wqh.DequeueTask)

	// TaskTemplate endpoints
	tth := handler.NewTaskTemplateHandler(mgr)
	api.POST("/tasktemplate", tth.CreateTaskTemplate)
	api.GET("/tasktemplate", tth.GetTaskTemplates)
	api.PUT("/tasktemplate/:id", tth.UpdateTaskTemplate)
	api.DELETE("/tasktemplate/:id", tth.DeleteTaskTemplate)

	return router, nil
}
