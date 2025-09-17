package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	taskforge "github.com/agincgit/taskforge"
	"github.com/agincgit/taskforge/model"
	"github.com/agincgit/taskforge/server"
)

func TestGetTaskByIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	router, err := server.NewRouter(db)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	seedTask := model.Task{
		Type:   "test-task",
		Status: string(taskforge.StatusPending),
	}

	if err := db.WithContext(context.Background()).Create(&seedTask).Error; err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/taskforge/api/v1/tasks/%s", seedTask.ID.String()), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var got model.Task
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if got.ID != seedTask.ID {
		t.Fatalf("expected task ID %s, got %s", seedTask.ID, got.ID)
	}
}
