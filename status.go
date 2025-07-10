package taskforge

// Status is the lifecycle state of a Task in TaskForge.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusFailed     Status = "failed"
	StatusComplete   Status = "complete"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusFailed, StatusComplete:
		return true
	}
	return false
}
