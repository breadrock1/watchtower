package domain

import "errors"

var (
	ErrExecution       = errors.New("execution error")
	ErrTaskNotFound    = errors.New("task not found")
	ErrInvalidTaskData = errors.New("invalid task data")
)
