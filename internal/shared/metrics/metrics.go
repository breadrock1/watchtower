package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	serviceLabel       = "service"
	statusLabel        = "status"
	targetLabelName    = "target"
	operationLabelName = "operation"
	methodLabelName    = "method"
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

func init() {
	RmqReconnectCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_rmq_reconnect_total",
			Help: "Total number of rmq reconnects",
		},
		[]string{serviceLabel},
	)

	UploadedFilesCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_upload_files_total",
			Help: "Total number of uploaded files to storage",
		},
		[]string{serviceLabel, statusLabel},
	)

	CreatedProcessingTasksCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_created_tasks_total",
			Help: "Total number of created tasks of processing",
		},
		[]string{serviceLabel, statusLabel},
	)

	ProcessingTasksStatusCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_processing_tasks_status_total",
			Help: "Total number of processing tasks with status",
		},
		[]string{serviceLabel, statusLabel},
	)

	PublishedTasksQueueCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_published_tasks_queue_total",
			Help: "Total number of published tasks to queue",
		},
		[]string{serviceLabel, "is_failed", "retries"},
	)

	OrchestratorProcessingCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_orchestrator_processed_total",
			Help: "Total processed documents into orchestrator",
		},
		[]string{serviceLabel, statusLabel},
	)

	OrchestratorProcessingDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_orchestrator_processing_duration_seconds",
			Help: "Latency of full document processing time in seconds",
		},
		[]string{serviceLabel, statusLabel},
	)

	RecognizerDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_recognizer_duration_seconds",
			Help: "Latency of recognizing text from document file",
		},
		[]string{serviceLabel, statusLabel},
	)

	StoreProcessedDocumentDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_store_document_duration_seconds",
			Help: "Latency of storing processed document",
		},
		[]string{serviceLabel, statusLabel},
	)

	S3OperationDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_s3_operation_duration_seconds",
			Help:    "Latency of general S3 operations (stat/copy/delete/multidelete/bucket) in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{serviceLabel, operationLabelName, statusLabel},
	)

	OutgoingHTTPRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_outgoing_http_request_duration_seconds",
			Help:    "Latency of outgoing HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{serviceLabel, targetLabelName, methodLabelName},
	)

	RedisOperationDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_redis_operation_duration_seconds",
			Help:    "Latency of Redis operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{serviceLabel, operationLabelName, statusLabel},
	)

	UploadFileSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_upload_file_size_bytes",
			Help:    "Size of uploaded files in bytes",
			Buckets: prometheus.ExponentialBuckets(1024, 4, 8),
		},
		[]string{serviceLabel},
	)

	OrchestratorAcquireWaitDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "watchtower_semaphore_acquire_wait_seconds",
			Help:    "Time spent waiting for semaphore in orchestrator worker pool",
			Buckets: prometheus.DefBuckets,
		},
		[]string{serviceLabel},
	)
}
