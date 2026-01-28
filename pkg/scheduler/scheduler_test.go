package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/agincgit/taskforge/pkg/model"
)

type stubManager struct {
	mu         sync.Mutex
	templates  []model.TaskTemplate
	createdIDs []uuid.UUID
}

func (s *stubManager) GetTaskTemplates(ctx context.Context) ([]model.TaskTemplate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	templates := make([]model.TaskTemplate, len(s.templates))
	copy(templates, s.templates)
	return templates, nil
}

func (s *stubManager) CreateTaskFromTemplate(ctx context.Context, templateID uuid.UUID, overrides map[string]interface{}, scheduledFor *time.Time) (*model.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.createdIDs = append(s.createdIDs, templateID)
	tplID := templateID
	return &model.Task{TemplateID: &tplID}, nil
}

func (s *stubManager) setTemplates(tpls []model.TaskTemplate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates = make([]model.TaskTemplate, len(tpls))
	copy(s.templates, tpls)
}

func (s *stubManager) createdCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.createdIDs)
}

func TestSchedulerStartRegistersRecurringTemplates(t *testing.T) {
	stub := &stubManager{}
	recurringID := uuid.New()
	stub.setTemplates([]model.TaskTemplate{{
		BaseModel:    model.BaseModel{ID: recurringID},
		IsRecurring:  true,
		CronSchedule: "*/5 * * * *",
	}, {
		BaseModel:   model.BaseModel{ID: uuid.New()},
		IsRecurring: false,
	}})

	sched := NewScheduler(stub)
	if err := sched.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer sched.cron.Stop()

	if len(sched.entries) != 1 {
		t.Fatalf("expected one scheduled entry, got %d", len(sched.entries))
	}

	entryID, ok := sched.entries[recurringID]
	if !ok {
		t.Fatalf("expected entry for template %s", recurringID)
	}

	// Execute the job directly to verify task creation.
	sched.cron.Entry(entryID).Job.Run()
	if stub.createdCount() != 1 {
		t.Fatalf("expected CreateTaskFromTemplate to be called once, got %d", stub.createdCount())
	}
}

func TestSchedulerReloadTemplatesRefreshesEntries(t *testing.T) {
	stub := &stubManager{}
	firstID := uuid.New()
	secondID := uuid.New()
	stub.setTemplates([]model.TaskTemplate{{
		BaseModel:    model.BaseModel{ID: firstID},
		IsRecurring:  true,
		CronSchedule: "0 * * * *",
	}})

	sched := NewScheduler(stub)
	if err := sched.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer sched.cron.Stop()

	if _, ok := sched.entries[firstID]; !ok {
		t.Fatalf("expected entry for first template")
	}

	stub.setTemplates([]model.TaskTemplate{{
		BaseModel:    model.BaseModel{ID: secondID},
		IsRecurring:  true,
		CronSchedule: "*/15 * * * *",
	}})

	if err := sched.ReloadTemplates(context.Background()); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	if len(sched.entries) != 1 {
		t.Fatalf("expected one entry after reload, got %d", len(sched.entries))
	}
	if _, ok := sched.entries[secondID]; !ok {
		t.Fatalf("expected entry for second template after reload")
	}
	if _, ok := sched.entries[firstID]; ok {
		t.Fatalf("did not expect entry for first template after reload")
	}
}

func TestOnTemplateChangedAddsAndRemovesEntries(t *testing.T) {
	stub := &stubManager{}
	tplID := uuid.New()
	stub.setTemplates(nil)

	sched := NewScheduler(stub)
	if err := sched.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer sched.cron.Stop()

	tpl := model.TaskTemplate{
		BaseModel:    model.BaseModel{ID: tplID},
		IsRecurring:  true,
		CronSchedule: "*/10 * * * *",
	}

	if err := sched.OnTemplateChanged(tpl); err != nil {
		t.Fatalf("unexpected error adding schedule: %v", err)
	}
	entryID, ok := sched.entries[tplID]
	if !ok {
		t.Fatalf("expected entry for template after change")
	}

	tpl.CronSchedule = "*/20 * * * *"
	if err := sched.OnTemplateChanged(tpl); err != nil {
		t.Fatalf("unexpected error updating schedule: %v", err)
	}
	newEntryID, ok := sched.entries[tplID]
	if !ok {
		t.Fatalf("expected entry after schedule update")
	}
	if entryID == newEntryID {
		t.Fatalf("expected entry ID to change after update")
	}

	tpl.IsRecurring = false
	if err := sched.OnTemplateChanged(tpl); err != nil {
		t.Fatalf("unexpected error disabling recurrence: %v", err)
	}
	if _, ok := sched.entries[tplID]; ok {
		t.Fatalf("expected entry to be removed when recurrence disabled")
	}

	sched.OnTemplateDeleted(tplID)
	if _, ok := sched.entries[tplID]; ok {
		t.Fatalf("expected no entry after deletion")
	}
}
