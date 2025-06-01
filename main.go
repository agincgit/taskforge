package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/agincgit/taskforge/config"
	"github.com/agincgit/taskforge/handler"
	"github.com/agincgit/taskforge/model"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.GetConfig("config.json")

	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate models
	db.AutoMigrate(&model.Task{}, &model.TaskInput{}, &model.TaskOutput{}, &model.TaskHistory{},
		&model.TaskTemplate{}, &model.WorkerType{}, &model.WorkerRegistration{},
		&model.WorkerHeartbeat{}, &model.DeadLetterQueue{}, &model.TaskCleanup{},
		&model.JobQueue{})

	// Setup handlers
	router := mux.NewRouter()
	wh := &handler.WorkerHandler{DB: db}
	wqh := &handler.WorkerQueueHandler{DB: db}
	tth := &handler.TaskTemplateHandler{DB: db}
	th := &handler.TaskHandler{DB: db}

	router.HandleFunc("/api/v1/worker", wh.RegisterWorker).Methods("POST")
	router.HandleFunc("/api/v1/worker/{id}/heartbeat", wh.Heartbeat).Methods("PUT")
	router.HandleFunc("/api/v1/workerqueue", wqh.EnqueueTask).Methods("POST")
	router.HandleFunc("/api/v1/workerqueue", wqh.GetQueue).Methods("GET")
	router.HandleFunc("/api/v1/workerqueue/{id}", wqh.DequeueTask).Methods("DELETE")
	router.HandleFunc("/api/v1/tasktemplate", tth.CreateTaskTemplate).Methods("POST")
	router.HandleFunc("/api/v1/tasktemplate", tth.GetTaskTemplates).Methods("GET")
	router.HandleFunc("/api/v1/tasktemplate/{id}", tth.UpdateTaskTemplate).Methods("PUT")
	router.HandleFunc("/api/v1/tasktemplate/{id}", tth.DeleteTaskTemplate).Methods("DELETE")
	router.HandleFunc("/api/v1/task", th.CreateTask).Methods("POST")
	router.HandleFunc("/api/v1/task", th.GetTasks).Methods("GET")
	router.HandleFunc("/api/v1/task/{id}", th.GetTask).Methods("GET")
	router.HandleFunc("/api/v1/task/{id}", th.UpdateTask).Methods("PUT")
	router.HandleFunc("/api/v1/task/{id}", th.DeleteTask).Methods("DELETE")

	fmt.Printf("Server starting on port %s...", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
