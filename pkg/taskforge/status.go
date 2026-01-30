package taskforge

// Status is the lifecycle state of a Task in TaskForge.
type Status string

const (
	StatusPending        Status = "pending"
	StatusInProgress     Status = "in_progress"
	StatusSucceeded      Status = "succeeded"
	StatusFailed         Status = "failed"
	StatusPendingCancel  Status = "pending_cancellation"
	StatusCancelled      Status = "cancelled"
	StatusFailedToCancel Status = "failed_to_cancel"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusSucceeded, StatusFailed,
		StatusPendingCancel, StatusCancelled, StatusFailedToCancel:
		return true
	}
	return false
}
