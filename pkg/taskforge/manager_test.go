package taskforge

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/internal/persistence"
	"github.com/agincgit/taskforge/pkg/model"
	"github.com/google/uuid"
)

func TestRetryTaskResetsFriendlyID(t *testing.T) {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := persistence.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	mgr, err := NewManager(Config{
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

	if err := persistence.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	mgr, err := NewManager(Config{
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

	if err := persistence.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	mgr, err := NewManager(Config{
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

func TestGetChildTasks(t *testing.T) {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := persistence.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	mgr, err := NewManager(Config{
		DB:        db,
		TableName: "tasks",
		Context:   ctx,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create parent task
	parent := model.Task{
		Type:   "workflow",
		Status: string(StatusInProgress),
	}
	if err := db.WithContext(ctx).Create(&parent).Error; err != nil {
		t.Fatalf("failed to create parent task: %v", err)
	}

	// Create child tasks
	child1 := model.Task{
		Type:         "step1",
		Status:       string(StatusPending),
		ParentTaskID: &parent.ID,
	}
	child2 := model.Task{
		Type:         "step2",
		Status:       string(StatusPending),
		ParentTaskID: &parent.ID,
	}
	if err := db.WithContext(ctx).Create(&child1).Error; err != nil {
		t.Fatalf("failed to create child1: %v", err)
	}
	if err := db.WithContext(ctx).Create(&child2).Error; err != nil {
		t.Fatalf("failed to create child2: %v", err)
	}

	// Create unrelated task (no parent)
	unrelated := model.Task{
		Type:   "other",
		Status: string(StatusPending),
	}
	if err := db.WithContext(ctx).Create(&unrelated).Error; err != nil {
		t.Fatalf("failed to create unrelated task: %v", err)
	}

	// Get children of parent
	children, err := mgr.GetChildTasks(ctx, parent.ID)
	if err != nil {
		t.Fatalf("failed to get child tasks: %v", err)
	}

	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	// Verify child IDs
	childIDs := make(map[uuid.UUID]bool)
	for _, c := range children {
		childIDs[c.ID] = true
	}
	if !childIDs[child1.ID] || !childIDs[child2.ID] {
		t.Fatalf("expected children to include child1 and child2")
	}
	if childIDs[unrelated.ID] {
		t.Fatalf("unrelated task should not be in children")
	}
}

func TestHasChildren(t *testing.T) {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := persistence.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	mgr, err := NewManager(Config{
		DB:        db,
		TableName: "tasks",
		Context:   ctx,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create parent with children
	parent := model.Task{
		Type:   "workflow",
		Status: string(StatusInProgress),
	}
	if err := db.WithContext(ctx).Create(&parent).Error; err != nil {
		t.Fatalf("failed to create parent task: %v", err)
	}

	child := model.Task{
		Type:         "step1",
		Status:       string(StatusPending),
		ParentTaskID: &parent.ID,
	}
	if err := db.WithContext(ctx).Create(&child).Error; err != nil {
		t.Fatalf("failed to create child: %v", err)
	}

	// Create task without children
	lonely := model.Task{
		Type:   "standalone",
		Status: string(StatusPending),
	}
	if err := db.WithContext(ctx).Create(&lonely).Error; err != nil {
		t.Fatalf("failed to create lonely task: %v", err)
	}

	// Test parent has children
	hasChildren, err := mgr.HasChildren(ctx, parent.ID)
	if err != nil {
		t.Fatalf("failed to check HasChildren: %v", err)
	}
	if !hasChildren {
		t.Fatalf("expected parent to have children")
	}

	// Test lonely task has no children
	hasChildren, err = mgr.HasChildren(ctx, lonely.ID)
	if err != nil {
		t.Fatalf("failed to check HasChildren: %v", err)
	}
	if hasChildren {
		t.Fatalf("expected lonely task to have no children")
	}
}

func TestGetTaskTree(t *testing.T) {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := persistence.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	mgr, err := NewManager(Config{
		DB:        db,
		TableName: "tasks",
		Context:   ctx,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create a tree structure:
	// root
	// ├── child1
	// │   └── grandchild1
	// └── child2

	root := model.Task{
		Type:   "root",
		Status: string(StatusInProgress),
	}
	if err := db.WithContext(ctx).Create(&root).Error; err != nil {
		t.Fatalf("failed to create root: %v", err)
	}

	child1 := model.Task{
		Type:         "child1",
		Status:       string(StatusPending),
		ParentTaskID: &root.ID,
	}
	if err := db.WithContext(ctx).Create(&child1).Error; err != nil {
		t.Fatalf("failed to create child1: %v", err)
	}

	child2 := model.Task{
		Type:         "child2",
		Status:       string(StatusPending),
		ParentTaskID: &root.ID,
	}
	if err := db.WithContext(ctx).Create(&child2).Error; err != nil {
		t.Fatalf("failed to create child2: %v", err)
	}

	grandchild1 := model.Task{
		Type:         "grandchild1",
		Status:       string(StatusPending),
		ParentTaskID: &child1.ID,
	}
	if err := db.WithContext(ctx).Create(&grandchild1).Error; err != nil {
		t.Fatalf("failed to create grandchild1: %v", err)
	}

	// Get task tree
	tree, err := mgr.GetTaskTree(ctx, root.ID)
	if err != nil {
		t.Fatalf("failed to get task tree: %v", err)
	}

	// Verify root
	if tree.Task.ID != root.ID {
		t.Fatalf("expected root ID %s, got %s", root.ID, tree.Task.ID)
	}

	// Verify root has 2 children
	if len(tree.Children) != 2 {
		t.Fatalf("expected root to have 2 children, got %d", len(tree.Children))
	}

	// Find child1 in tree and verify it has grandchild
	var child1Node *TaskNode
	for i := range tree.Children {
		if tree.Children[i].Task.ID == child1.ID {
			child1Node = &tree.Children[i]
			break
		}
	}
	if child1Node == nil {
		t.Fatalf("child1 not found in tree")
	}
	if len(child1Node.Children) != 1 {
		t.Fatalf("expected child1 to have 1 child, got %d", len(child1Node.Children))
	}
	if child1Node.Children[0].Task.ID != grandchild1.ID {
		t.Fatalf("expected grandchild1 ID %s, got %s", grandchild1.ID, child1Node.Children[0].Task.ID)
	}

	// Verify child2 has no children
	var child2Node *TaskNode
	for i := range tree.Children {
		if tree.Children[i].Task.ID == child2.ID {
			child2Node = &tree.Children[i]
			break
		}
	}
	if child2Node == nil {
		t.Fatalf("child2 not found in tree")
	}
	if len(child2Node.Children) != 0 {
		t.Fatalf("expected child2 to have 0 children, got %d", len(child2Node.Children))
	}
}
