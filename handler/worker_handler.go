package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
)

type WorkerHandler struct {
	DB *gorm.DB
}

func (h *WorkerHandler) RegisterWorker(w http.ResponseWriter, r *http.Request) {
	var reg model.WorkerRegistration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.DB.Create(&reg).Error; err != nil {
		http.Error(w, "Failed to register worker", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reg)
}

func (h *WorkerHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var beat model.WorkerHeartbeat
	if err := h.DB.First(&beat, "worker_id = ?", id).Error; err != nil {
		http.Error(w, "Worker not found", http.StatusNotFound)
		return
	}
	beat.LastPing = time.Now()
	h.DB.Save(&beat)
	w.WriteHeader(http.StatusOK)
}
