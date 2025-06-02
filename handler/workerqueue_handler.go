package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
)

// WorkerQueueHandler manages worker queue operations.
type WorkerQueueHandler struct {
	DB *gorm.DB
}

// NewWorkerQueueHandler constructs a WorkerQueueHandler.
func NewWorkerQueueHandler(db *gorm.DB) *WorkerQueueHandler {
	return &WorkerQueueHandler{DB: db}
}

// EnqueueTask adds a new job to the queue.
func (h *WorkerQueueHandler) EnqueueTask(w http.ResponseWriter, r *http.Request) {
	var jq model.JobQueue
	if err := json.NewDecoder(r.Body).Decode(&jq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.DB.Create(&jq).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(jq)
}

// GetQueue returns all queued jobs.
func (h *WorkerQueueHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	var queue []model.JobQueue
	if err := h.DB.Find(&queue).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(queue)
}

// DequeueTask removes a job from the queue by its ID.
func (h *WorkerQueueHandler) DequeueTask(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		http.Error(w, "invalid ID parameter", http.StatusBadRequest)
		return
	}
	if err := h.DB.Delete(&model.JobQueue{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
