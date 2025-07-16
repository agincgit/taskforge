package taskforge

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/handler"
)

// NewRouter sets up database migrations and registers all TaskForge API routes.
func NewRouter(db *gorm.DB) (*gin.Engine, error) {
	_, err := DBMigrate(db)
	if err != nil {
		fmt.Println("Database changes failed to apply")
	}
	router := gin.Default()
	api := router.Group("/taskforge/api/v1")

	// Task endpoints
	th := handler.NewTaskHandler(db)
	api.POST("/tasks", th.CreateTask)
	api.GET("/tasks", th.GetTasks)
	api.PUT("/tasks/:id", th.UpdateTask)
	api.DELETE("/tasks/:id", th.DeleteTask)

	// WorkerQueue endpoints
	wqh := handler.NewWorkerQueueHandler(db)
	api.POST("/workerqueue", wqh.EnqueueTask)
	api.GET("/workerqueue", wqh.GetQueue)
	api.DELETE("/workerqueue/:id", wqh.DequeueTask)

	// TaskTemplate endpoints
	tth := handler.NewTaskTemplateHandler(db)
	api.POST("/tasktemplate", tth.CreateTaskTemplate)
	api.GET("/tasktemplate", tth.GetTaskTemplates)
	api.PUT("/tasktemplate/:id", tth.UpdateTaskTemplate)
	api.DELETE("/tasktemplate/:id", tth.DeleteTaskTemplate)

	return router, nil
}
