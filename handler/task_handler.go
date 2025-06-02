package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
)

type TaskHandler struct {
	DB *gorm.DB
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{DB: db}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var t model.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	h.DB.Create(&t)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []model.Task
	h.DB.Find(&tasks)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var t model.Task
	if err := h.DB.First(&t, "id = ?", id).Error; err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(t)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var t model.Task
	if h.DB.First(&t, "id = ?", id).Error != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	json.NewDecoder(r.Body).Decode(&t)
	h.DB.Save(&t)
	json.NewEncoder(w).Encode(t)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	h.DB.Delete(&model.Task{}, "id = ?", id)
	w.WriteHeader(http.StatusNoContent)
}
