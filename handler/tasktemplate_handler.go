package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
)

type TaskTemplateHandler struct {
	DB *gorm.DB
}

func NewTaskTemplateHandler(db *gorm.DB) *TaskTemplateHandler {
	return &TaskTemplateHandler{DB: db}
}

func (h *TaskTemplateHandler) CreateTaskTemplate(w http.ResponseWriter, r *http.Request) {
	var tmpl model.TaskTemplate
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.DB.Create(&tmpl).Error; err != nil {
		http.Error(w, "Failed to create template", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tmpl)
}

func (h *TaskTemplateHandler) GetTaskTemplates(w http.ResponseWriter, r *http.Request) {
	var tpls []model.TaskTemplate
	h.DB.Find(&tpls)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tpls)
}

func (h *TaskTemplateHandler) UpdateTaskTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var tmpl model.TaskTemplate
	if err := h.DB.First(&tmpl, "id = ?", id).Error; err != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}
	json.NewDecoder(r.Body).Decode(&tmpl)
	h.DB.Save(&tmpl)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tmpl)
}

func (h *TaskTemplateHandler) DeleteTaskTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	h.DB.Delete(&model.TaskTemplate{}, "id = ?", id)
	w.WriteHeader(http.StatusNoContent)
}
