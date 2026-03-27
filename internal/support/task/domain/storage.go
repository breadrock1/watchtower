package domain

import (
	"watchtower/internal/shared/kernel"
)

// ITaskStorage defines the interface for persistent task storage.
// This is used to track task state independently from the message queue.
type ITaskStorage interface {
	ITaskManager
}

// ITaskManager defines operations for managing task lifecycle in persistent storage.
// Tasks are stored independently of the queue to maintain state across system restarts
// and provide audit capabilities.
type ITaskManager interface {
	// GetTask retrieves a task by its bucket and task IDs.
	// This is useful for checking task status or retrieving results.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket containing the task's input
	//   - taskID: Unique identifier of the task
	//
	// Returns:
	//   - *Task: Complete task information including current status
	//   - error: ErrExecution if returned operation error,
	//			  ErrTaskNotFound if task not found,
	//			  ErrInvalidTaskData if update would violate constraints,
	//  		  or other storage errors
	//
	// Example:
	//   task, err := storage.GetTask(ctx, "input-bucket", taskID)
	//   if err == nil {
	//       fmt.Printf("Task %s status: %s\n", task.ID, task.Status)
	//   }
	GetTask(ctx kernel.Ctx, bucketID kernel.BucketID, taskID kernel.TaskID) (*Task, error)

	// GetAllBucketTasks retrieves all tasks associated with a specific bucket.
	// This is useful for monitoring and batch operations.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket to get tasks for
	//
	// Returns:
	//   - []*Task: Slice of all tasks for the bucket, ordered by creation time
	//   - error: ErrExecution if returned operation error,
	//			  ErrInvalidTaskData if update would violate constraints,
	//            or other storage errors
	//
	// Example:
	//   tasks, err := storage.GetAllBucketTasks(ctx, "input-bucket")
	//   for _, task := range tasks {
	//       fmt.Printf("Task %s: %s\n", task.ID, task.Status)
	//   }
	GetAllBucketTasks(ctx kernel.Ctx, bucketID kernel.BucketID) ([]*Task, error)

	// UpdateTask updates an existing task's status and metadata.
	// This is called as tasks progress through their lifecycle.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - task: Complete task object with updated fields
	//
	// Returns:
	//   - error: ErrExecution if returned operation error,
	//			  ErrTaskNotFound if task not found,
	//            ErrInvalidTaskData if update would violate constraints,
	//            or other storage errors
	//
	// Example:
	//   task.Status = Processing
	//   task.StatusText = "Starting processing..."
	//   task.ModifiedAt = time.Now()
	//
	//   err := storage.UpdateTask(ctx, task)
	//   if err != nil {
	//       log.Printf("Failed to update task: %v", err)
	//   }
	UpdateTask(ctx kernel.Ctx, task *Task) error
}
