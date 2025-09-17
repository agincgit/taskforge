package taskforge

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
)

func TestRetryTaskResetsFriendlyID(t *testing.T) {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	mgr, err := NewManager(TaskForgeConfig{
		DB:        db,
		TableName: "tasks",
		Context:   ctx,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	original := model.Task{
		Type:    "retryable",
		Status:  string(StatusFailed),
		Attempt: 1,
	}

	if err := db.WithContext(ctx).Create(&original).Error; err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	if original.FriendlyID == 0 {
		t.Fatalf("expected original friendly id to be set")
	}

	retryTask, err := mgr.RetryTask(ctx, original.ID)
	if err != nil {
		t.Fatalf("retry task failed: %v", err)
	}

	if retryTask.FriendlyID == original.FriendlyID {
		t.Fatalf("expected new friendly id to differ from original: %d", retryTask.FriendlyID)
	}

	var stored model.Task
	if err := db.WithContext(ctx).First(&stored, "id = ?", retryTask.ID).Error; err != nil {
		t.Fatalf("failed to load stored retry task: %v", err)
	}

	if stored.FriendlyID == original.FriendlyID {
		t.Fatalf("expected stored friendly id to differ from original: %d", stored.FriendlyID)
	}

	if stored.FriendlyID != retryTask.FriendlyID {
		t.Fatalf("expected stored friendly id to match returned friendly id: got %d, want %d", stored.FriendlyID, retryTask.FriendlyID)
	}
}
