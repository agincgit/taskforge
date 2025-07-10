package taskforge

import (
	"fmt"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/handler"
)

// NewRouter sets up database migrations and registers all TaskForge API routes.
func NewRouter(db *gorm.DB) (*mux.Router, error) {
	/* Auto-migrate all models
	if err := db.AutoMigrate(
		&model.Task{},
		&model.TaskInput{},
		&model.TaskOutput{},
		&model.TaskHistory{},
		&model.TaskTemplate{},
		&model.WorkerType{},
		&model.WorkerRegistration{},
		&model.WorkerHeartbeat{},
		&model.DeadLetterQueue{},
		&model.TaskCleanup{},
		&model.JobQueue{},
	); err != nil {
		return nil, fmt.Errorf("auto-migrate failed: %v", err)
	}*/
	_, err := DBMigrate(db)
	if err != nil {
		fmt.Println("Database changes failed to apply")
	}
	router := mux.NewRouter()
	api := router.PathPrefix("/taskforge/api/v1").Subrouter()

	// Task endpoints
	th := handler.NewTaskHandler(db)
	api.HandleFunc("/tasks", th.CreateTask).Methods("POST")
	api.HandleFunc("/tasks", th.GetTasks).Methods("GET")
	api.HandleFunc("/tasks/{id}", th.UpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", th.DeleteTask).Methods("DELETE")

	// WorkerQueue endpoints
	wqh := handler.NewWorkerQueueHandler(db)
	api.HandleFunc("/workerqueue", wqh.EnqueueTask).Methods("POST")
	api.HandleFunc("/workerqueue", wqh.GetQueue).Methods("GET")
	api.HandleFunc("/workerqueue/{id}", wqh.DequeueTask).Methods("DELETE")

	// TaskTemplate endpoints
	tth := handler.NewTaskTemplateHandler(db)
	api.HandleFunc("/tasktemplate", tth.CreateTaskTemplate).Methods("POST")
	api.HandleFunc("/tasktemplate", tth.GetTaskTemplates).Methods("GET")
	api.HandleFunc("/tasktemplate/{id}", tth.UpdateTaskTemplate).Methods("PUT")
	api.HandleFunc("/tasktemplate/{id}", tth.DeleteTaskTemplate).Methods("DELETE")

	return router, nil
}
