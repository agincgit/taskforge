// Package server provides HTTP server setup and routing for TaskForge.
package server

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/internal/handlers"
	"github.com/agincgit/taskforge/internal/persistence"
	"github.com/agincgit/taskforge/pkg/scheduler"
	"github.com/agincgit/taskforge/pkg/taskforge"
)

// NewRouter sets up database migrations and registers all TaskForge API routes.
func NewRouter(db *gorm.DB) (*gin.Engine, error) {
	// Run migrations
	if err := persistence.Migrate(db); err != nil {
		return nil, err
	}

	mgr, err := taskforge.NewManager(taskforge.Config{
		DB:        db,
		TableName: "tasks",
		Context:   context.Background(),
	})
	if err != nil {
		return nil, err
	}

	sched := scheduler.NewScheduler(mgr)
	if err := sched.Start(context.Background()); err != nil {
		return nil, err
	}

	router := gin.Default()
	api := router.Group("/taskforge/api/v1")

	// Task endpoints
	th := handlers.NewTaskHandler(mgr)
	api.POST("/tasks", th.CreateTask)
	api.GET("/tasks", th.GetTasks)
	api.GET("/tasks/:id", th.GetTask)
	api.PUT("/tasks/:id", th.UpdateTask)
	api.DELETE("/tasks/:id", th.DeleteTask)

	// WorkerQueue endpoints
	wqh := handlers.NewWorkerQueueHandler(mgr)
	api.POST("/workerqueue", wqh.EnqueueTask)
	api.GET("/workerqueue", wqh.GetQueue)
	api.DELETE("/workerqueue/:id", wqh.DequeueTask)

	// TaskTemplate endpoints
	tth := handlers.NewTaskTemplateHandler(mgr, sched)
	api.POST("/tasktemplate", tth.CreateTaskTemplate)
	api.GET("/tasktemplate", tth.GetTaskTemplates)
	api.PUT("/tasktemplate/:id", tth.UpdateTaskTemplate)
	api.DELETE("/tasktemplate/:id", tth.DeleteTaskTemplate)

	// Worker endpoints
	wh := handlers.NewWorkerHandler(mgr)
	api.POST("/workers", wh.RegisterWorker)
	api.PUT("/workers/:id/heartbeat", wh.Heartbeat)

	return router, nil
}
