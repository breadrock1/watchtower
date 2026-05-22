package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RmqReconnectCounter           *prometheus.CounterVec
	UploadedFilesCounter          *prometheus.CounterVec
	CreatedProcessingTasksCounter *prometheus.CounterVec
	OrchestratorProcessingCounter *prometheus.CounterVec
	ProcessingTasksStatusCounter  *prometheus.CounterVec
	PublishedTasksQueueCounter    *prometheus.CounterVec

	OrchestratorProcessingDurationSeconds *prometheus.HistogramVec
	RecognizerDurationSeconds             *prometheus.HistogramVec
	StoreProcessedDocumentDurationSeconds *prometheus.HistogramVec
)

const (
	SERVICE_LABEL_NAME   = "service"
	IS_FAILED_LABEL_NAME = "is_failed"
	STATUS_LABEL_NAME    = "status"
)

func init() {
	RmqReconnectCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_rmq_reconnect_total",
			Help: "Total number of rmq reconnects",
		},
		[]string{SERVICE_LABEL_NAME, IS_FAILED_LABEL_NAME},
	)

	UploadedFilesCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_upload_files_total",
			Help: "Total number of uploaded files to storage",
		},
		[]string{SERVICE_LABEL_NAME, IS_FAILED_LABEL_NAME},
	)

	CreatedProcessingTasksCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_created_tasks_total",
			Help: "Total number of created tasks of processing",
		},
		[]string{SERVICE_LABEL_NAME, IS_FAILED_LABEL_NAME},
	)

	ProcessingTasksStatusCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_processing_tasks_status_total",
			Help: "Total number of processing tasks with status",
		},
		[]string{SERVICE_LABEL_NAME, STATUS_LABEL_NAME},
	)

	PublishedTasksQueueCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_published_tasks_queue_total",
			Help: "Total number of published tasks to queue",
		},
		[]string{SERVICE_LABEL_NAME, "is_failed", "retries"},
	)

	OrchestratorProcessingCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_orchestrator_processed_total",
			Help: "Total processed documents into orchestrator",
		},
		[]string{SERVICE_LABEL_NAME, STATUS_LABEL_NAME},
	)

	OrchestratorProcessingDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_orchestrator_processing_duration_seconds",
			Help: "Latency of full document processing time in seconds",
		},
		[]string{SERVICE_LABEL_NAME, STATUS_LABEL_NAME},
	)

	RecognizerDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_recognizer_duration_seconds",
			Help: "Latency of recognizing text from document file",
		},
		[]string{SERVICE_LABEL_NAME, IS_FAILED_LABEL_NAME},
	)

	StoreProcessedDocumentDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_store_document_duration_seconds",
			Help: "Latency of storing processed document",
		},
		[]string{SERVICE_LABEL_NAME, IS_FAILED_LABEL_NAME},
	)
}
