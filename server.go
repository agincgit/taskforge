// server.go
package taskforge

import (
	"fmt"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/handler"
	"github.com/agincgit/taskforge/model"
)

func NewRouter(db *gorm.DB) (*mux.Router, error) {
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
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/taskforge/api/v1").Subrouter()

	// Tasks
	th := handler.NewTaskHandler(db)
	api.HandleFunc("/tasks", th.CreateTask).Methods("POST")
	api.HandleFunc("/tasks", th.GetTasks).Methods("GET")
	api.HandleFunc("/tasks/{id}", th.UpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", th.DeleteTask).Methods("DELETE")

	// WorkerQueue
	wqh := handler.NewWorkerQueueHandler(db)
	api.HandleFunc("/workerqueue", wqh.EnqueueTask).Methods("POST")
	api.HandleFunc("/workerqueue", wqh.GetQueue).Methods("GET")
	api.HandleFunc("/workerqueue/{id}", wqh.DequeueTask).Methods("DELETE")

	// TaskTemplates
	tth := handler.NewTaskTemplateHandler(db)
	api.HandleFunc("/tasktemplate", tth.CreateTaskTemplate).Methods("POST")
	api.HandleFunc("/tasktemplate", tth.GetTaskTemplates).Methods("GET")
	api.HandleFunc("/tasktemplate/{id}", tth.UpdateTaskTemplate).Methods("PUT")
	api.HandleFunc("/tasktemplate/{id}", tth.DeleteTaskTemplate).Methods("DELETE")

	return r, nil
}
