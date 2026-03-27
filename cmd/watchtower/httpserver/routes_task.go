package httpserver

import (
	"github.com/gofiber/fiber/v2"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slices"

	"watchtower/cmd/watchtower/httpserver/form"

	task "watchtower/internal/support/task/domain"
)

func (s *Server) CreateTasksGroup(group fiber.Router) {
	group.Get("/tasks/:bucket", s.LoadTasks)
	group.Get("/tasks/:bucket/:task_id", s.LoadTaskByID)
}

// LoadTasks
// @Summary Load processing tasks of uploaded files into bucket
// @Description Load tasks (processing/unrecognized/done) of uploaded files
// @ID load-tasks
// @Tags tasks
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket id of uploaded files"
// @Param status query string false "Status tasks to filter target result"
// @Success 200 {object} []form.TaskSchema "Loaded tasks"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/tasks/{bucket} [get]
func (s *Server) LoadTasks(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	status, err := ExtractTaskStatusParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString("unknown status")
	}

	span.SetAttributes(attribute.Int("status", status))

	taskStorage := s.state.GetTaskProcessor()
	tasks, err := taskStorage.GetBucketTasks(ctx, bucket)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	taskStatus := task.TaskStatus(status)
	foundedTasks := slices.DeleteFunc(tasks, func(task *task.Task) bool {
		return task.Status != taskStatus
	})

	foundedTasksDto := make([]form.TaskSchema, len(foundedTasks))
	for index, taskIt := range foundedTasks {
		foundedTasksDto[index] = form.TaskFromDomain(*taskIt)
	}

	return eCtx.Status(fiber.StatusOK).JSON(foundedTasksDto)
}

// LoadTaskByID
// @Summary Load processing task by id
// @Description Load processing/unrecognized/done task by id of uploaded file
// @ID load-task-by-id
// @Tags tasks
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket id of processing task"
// @Param task_id path string true "Task ID"
// @Success 200 {object} form.TaskSchema "Loaded tasks"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Task not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/tasks/{bucket}/{task_id} [get]
func (s *Server) LoadTaskByID(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	taskID, err := ExtractTaskIDParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	taskStorage := s.state.GetTaskProcessor()
	foundedTask, err := taskStorage.GetTask(ctx, bucket, taskID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	taskSchema := form.TaskFromDomain(*foundedTask)
	return eCtx.Status(fiber.StatusOK).JSON(taskSchema)
}
