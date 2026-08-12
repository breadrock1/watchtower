package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	UploadedFilesCounter                   *prometheus.CounterVec
	CreatedProcessingTasksCounter          *prometheus.CounterVec
	OrchestratorProcessingCounter          *prometheus.CounterVec
	OrchestratorAcquireWaitDurationSeconds *prometheus.HistogramVec
	ProcessingTasksStatusCounter           *prometheus.CounterVec
	PublishedTasksQueueCounter             *prometheus.CounterVec

	OrchestratorProcessingDurationSeconds *prometheus.HistogramVec
	RecognizerDurationSeconds             *prometheus.HistogramVec
	StoreProcessedDocumentDurationSeconds *prometheus.HistogramVec

	RedisOperationDurationSeconds *prometheus.HistogramVec
	RmqReconnectCounter           *prometheus.CounterVec
	S3OperationDurationSeconds    *prometheus.HistogramVec

	OutgoingHTTPRequestDurationSeconds *prometheus.HistogramVec

	UploadFileSizeBytes *prometheus.HistogramVec
)

const (
	SERVICE_LABEL        = "service"
	STATUS_LABEL         = "status"
	TARGET_LABEL_NAME    = "target"
	OPERATION_LABEL_NAME = "operation"
	METHOD_LABEL_NAME    = "method"
)

func init() {
	RmqReconnectCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_rmq_reconnect_total",
			Help: "Total number of rmq reconnects",
		},
		[]string{SERVICE_LABEL},
	)

	UploadedFilesCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_upload_files_total",
			Help: "Total number of uploaded files to storage",
		},
		[]string{SERVICE_LABEL, STATUS_LABEL},
	)

	CreatedProcessingTasksCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_created_tasks_total",
			Help: "Total number of created tasks of processing",
		},
		[]string{SERVICE_LABEL, STATUS_LABEL},
	)

	ProcessingTasksStatusCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_processing_tasks_status_total",
			Help: "Total number of processing tasks with status",
		},
		[]string{SERVICE_LABEL, STATUS_LABEL},
	)

	PublishedTasksQueueCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_published_tasks_queue_total",
			Help: "Total number of published tasks to queue",
		},
		[]string{SERVICE_LABEL, "is_failed", "retries"},
	)

	OrchestratorProcessingCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_orchestrator_processed_total",
			Help: "Total processed documents into orchestrator",
		},
		[]string{SERVICE_LABEL, STATUS_LABEL},
	)

	OrchestratorProcessingDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_orchestrator_processing_duration_seconds",
			Help: "Latency of full document processing time in seconds",
		},
		[]string{SERVICE_LABEL, STATUS_LABEL},
	)

	RecognizerDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_recognizer_duration_seconds",
			Help: "Latency of recognizing text from document file",
		},
		[]string{SERVICE_LABEL, STATUS_LABEL},
	)

	StoreProcessedDocumentDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_store_document_duration_seconds",
			Help: "Latency of storing processed document",
		},
		[]string{SERVICE_LABEL, STATUS_LABEL},
	)

	S3OperationDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_s3_operation_duration_seconds",
			Help:    "Latency of general S3 operations (stat/copy/delete/multidelete/bucket) in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{SERVICE_LABEL, OPERATION_LABEL_NAME, STATUS_LABEL},
	)

	OutgoingHTTPRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_outgoing_http_request_duration_seconds",
			Help:    "Latency of outgoing HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{SERVICE_LABEL, TARGET_LABEL_NAME, METHOD_LABEL_NAME},
	)

	RedisOperationDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_redis_operation_duration_seconds",
			Help:    "Latency of Redis operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{SERVICE_LABEL, OPERATION_LABEL_NAME},
	)

	UploadFileSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_upload_file_size_bytes",
			Help:    "Size of uploaded files in bytes",
			Buckets: prometheus.ExponentialBuckets(1024, 4, 8),
		},
		[]string{SERVICE_LABEL},
	)

	OrchestratorAcquireWaitDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_semaphore_acquire_wait_seconds",
			Help:    "Time spent waiting for semaphore in orchestrator worker pool",
			Buckets: prometheus.DefBuckets,
		},
		[]string{SERVICE_LABEL},
	)
}
