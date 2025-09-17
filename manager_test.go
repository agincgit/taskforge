package taskforge

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/model"
	"github.com/google/uuid"
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

func TestCreateTaskFromTemplate(t *testing.T) {
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

	worker := model.WorkerType{
		Name:        "emailer",
		Description: "handles email jobs",
	}
	if err := db.WithContext(ctx).Create(&worker).Error; err != nil {
		t.Fatalf("failed to seed worker type: %v", err)
	}

	defaultInputs := map[string]interface{}{
		"subject": "hello",
		"retry":   3,
	}
	defaultBytes, err := json.Marshal(defaultInputs)
	if err != nil {
		t.Fatalf("failed to marshal default inputs: %v", err)
	}

	tmpl := model.TaskTemplate{
		Name:           "send_email",
		Description:    "Send notification email",
		WorkerTypeID:   worker.ID,
		IsRecurring:    false,
		ExpirationTime: time.Hour,
		DefaultInputs:  string(defaultBytes),
	}
	if err := db.WithContext(ctx).Create(&tmpl).Error; err != nil {
		t.Fatalf("failed to seed template: %v", err)
	}

	overrides := map[string]interface{}{
		"retry":     5,
		"recipient": "ops@example.com",
	}
	scheduledFor := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)

	task, err := mgr.CreateTaskFromTemplate(ctx, tmpl.ID, overrides, &scheduledFor)
	if err != nil {
		t.Fatalf("failed to create task from template: %v", err)
	}

	if task.TemplateID == nil || *task.TemplateID != tmpl.ID {
		t.Fatalf("expected task template id %s, got %v", tmpl.ID, task.TemplateID)
	}
	if task.ScheduledFor == nil || !task.ScheduledFor.Equal(scheduledFor) {
		t.Fatalf("expected scheduled time %s, got %v", scheduledFor, task.ScheduledFor)
	}
	if task.Type != worker.Name {
		t.Fatalf("expected task type %q, got %q", worker.Name, task.Type)
	}
	if task.Status != string(StatusPending) {
		t.Fatalf("expected task status %q, got %q", StatusPending, task.Status)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if len(payload) != 3 {
		t.Fatalf("expected merged payload to contain 3 keys, got %d", len(payload))
	}
	if got, ok := payload["subject"].(string); !ok || got != "hello" {
		t.Fatalf("expected payload subject 'hello', got %#v", payload["subject"])
	}
	if got, ok := payload["recipient"].(string); !ok || got != "ops@example.com" {
		t.Fatalf("expected payload recipient 'ops@example.com', got %#v", payload["recipient"])
	}
	if got, ok := payload["retry"].(float64); !ok || got != 5 {
		t.Fatalf("expected payload retry 5, got %#v", payload["retry"])
	}

	var stored model.Task
	if err := db.WithContext(ctx).First(&stored, "id = ?", task.ID).Error; err != nil {
		t.Fatalf("failed to load stored task: %v", err)
	}

	if stored.TemplateID == nil || *stored.TemplateID != tmpl.ID {
		t.Fatalf("expected stored task template id %s, got %v", tmpl.ID, stored.TemplateID)
	}
	if stored.ScheduledFor == nil || !stored.ScheduledFor.Equal(scheduledFor) {
		t.Fatalf("expected stored scheduled time %s, got %v", scheduledFor, stored.ScheduledFor)
	}
	if stored.Type != worker.Name {
		t.Fatalf("expected stored task type %q, got %q", worker.Name, stored.Type)
	}
	if stored.Status != string(StatusPending) {
		t.Fatalf("expected stored task status %q, got %q", StatusPending, stored.Status)
	}

	var inputs []model.TaskInput
	if err := db.WithContext(ctx).Where("task_id = ?", task.FriendlyID).Find(&inputs).Error; err != nil {
		t.Fatalf("failed to load task inputs: %v", err)
	}
	if len(inputs) != 3 {
		t.Fatalf("expected 3 task inputs, got %d", len(inputs))
	}

	values := make(map[string]interface{})
	for _, in := range inputs {
		var v interface{}
		if err := json.Unmarshal([]byte(in.InputValue), &v); err != nil {
			t.Fatalf("failed to decode input %s: %v", in.InputKey, err)
		}
		values[in.InputKey] = v
	}

	if got, ok := values["subject"].(string); !ok || got != "hello" {
		t.Fatalf("expected input subject 'hello', got %#v", values["subject"])
	}
	if got, ok := values["recipient"].(string); !ok || got != "ops@example.com" {
		t.Fatalf("expected input recipient 'ops@example.com', got %#v", values["recipient"])
	}
	if got, ok := values["retry"].(float64); !ok || got != 5 {
		t.Fatalf("expected input retry 5, got %#v", values["retry"])
	}
}

func TestCreateTaskFromTemplateMissingTemplate(t *testing.T) {
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

	_, err = mgr.CreateTaskFromTemplate(ctx, uuid.New(), nil, nil)
	if err == nil {
		t.Fatalf("expected error when template is missing")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}
