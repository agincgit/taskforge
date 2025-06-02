package handler

import (
	"encoding/json"
	"net/http"

	"github.com/agincgit/taskforge/model"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type WorkerQueueHandler struct {
	DB *gorm.DB
}
func NewWorkerQueueHandler(db *gorm.DB) *WorkerQueueHandler {
    return &WorkerQueueHandler{DB: db}
}\
func (h *WorkerQueueHandler) EnqueueTask(w http.ResponseWriter, r *http.Request) {
	var entry model.JobQueue
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.DB.Create(&entry).Error; err != nil {
		http.Error(w, "Failed to enqueue task", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

func (h *WorkerQueueHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	var queue []model.JobQueue
	workerID := r.URL.Query().Get("worker_id")
	query := h.DB
	if workerID != "" {
		query = query.Where("worker_id = ?", workerID)
	}
	query.Find(&queue)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(queue)
}

func (h *WorkerQueueHandler) DequeueTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	h.DB.Delete(&model.JobQueue{}, "task_id = ?", id)
	w.WriteHeader(http.StatusNoContent)
}
