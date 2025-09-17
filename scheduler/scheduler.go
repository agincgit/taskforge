package scheduler

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"github.com/agincgit/taskforge/model"
)

// Logger is a minimal logging interface compatible with taskforge.Manager.
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Manager defines the subset of manager capabilities the Scheduler relies on.
type Manager interface {
	GetTaskTemplates(ctx context.Context) ([]model.TaskTemplate, error)
	CreateTaskFromTemplate(ctx context.Context, templateID uuid.UUID, overrides map[string]interface{}, scheduledFor *time.Time) (*model.Task, error)
}

// Option configures optional Scheduler behaviors.
type Option func(*Scheduler)

// WithLogger installs a logger used for informational and error logs.
func WithLogger(l Logger) Option {
	return func(s *Scheduler) {
		s.logger = l
	}
}

// Scheduler manages cron registrations for recurring task templates.
type Scheduler struct {
	mgr     Manager
	cron    *cron.Cron
	entries map[uuid.UUID]cron.EntryID
	logger  Logger

	mu      sync.RWMutex
	ctx     context.Context
	started bool
}

// NewScheduler constructs a Scheduler backed by the provided manager.
func NewScheduler(mgr Manager, opts ...Option) *Scheduler {
	s := &Scheduler{
		mgr:     mgr,
		cron:    cron.New(),
		entries: make(map[uuid.UUID]cron.EntryID),
		ctx:     context.Background(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Start bootstraps cron registrations from the manager and launches the cron runner.
func (s *Scheduler) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.Lock()
	s.ctx = ctx
	alreadyStarted := s.started
	s.mu.Unlock()

	if err := s.ReloadTemplates(ctx); err != nil {
		return err
	}

	if !alreadyStarted {
		s.mu.Lock()
		if !s.started {
			s.cron.Start()
			s.started = true
		}
		s.mu.Unlock()
	}
	return nil
}

// ReloadTemplates clears and re-registers cron jobs using the latest templates.
func (s *Scheduler) ReloadTemplates(ctx context.Context) error {
	ctx = s.getContext(ctx)
	templates, err := s.mgr.GetTaskTemplates(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entryID := range s.entries {
		s.cron.Remove(entryID)
	}
	s.entries = make(map[uuid.UUID]cron.EntryID)

	for _, tpl := range templates {
		if !shouldSchedule(tpl) {
			continue
		}
		if err := s.addTemplateLocked(tpl); err != nil {
			return err
		}
	}
	return nil
}

// OnTemplateChanged reconciles cron entries for a specific template after creation or update.
func (s *Scheduler) OnTemplateChanged(tpl model.TaskTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.entries[tpl.ID]; ok {
		s.cron.Remove(entryID)
		delete(s.entries, tpl.ID)
	}

	if !shouldSchedule(tpl) {
		return nil
	}

	return s.addTemplateLocked(tpl)
}

// OnTemplateDeleted removes any scheduled entry associated with the template.
func (s *Scheduler) OnTemplateDeleted(templateID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeLocked(templateID)
}

func (s *Scheduler) addTemplateLocked(tpl model.TaskTemplate) error {
	tplCopy := tpl
	entryID, err := s.cron.AddFunc(tplCopy.CronSchedule, func() {
		scheduledFor := time.Now()
		if _, err := s.mgr.CreateTaskFromTemplate(s.getContext(nil), tplCopy.ID, nil, &scheduledFor); err != nil {
			s.logError("scheduler: failed to create task from template %s: %v", tplCopy.ID, err)
		}
	})
	if err != nil {
		return err
	}
	s.entries[tpl.ID] = entryID
	return nil
}

func (s *Scheduler) removeLocked(templateID uuid.UUID) {
	if entryID, ok := s.entries[templateID]; ok {
		s.cron.Remove(entryID)
		delete(s.entries, templateID)
	}
}

func (s *Scheduler) getContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

func (s *Scheduler) logError(format string, args ...interface{}) {
	if s.logger != nil {
		s.logger.Errorf(format, args...)
	}
}

func shouldSchedule(tpl model.TaskTemplate) bool {
	return tpl.IsRecurring && strings.TrimSpace(tpl.CronSchedule) != ""
}
