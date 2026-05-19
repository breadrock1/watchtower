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

func init() {
	RmqReconnectCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_rmq_reconnect_total",
			Help: "Total number of rmq reconnects",
		},
		[]string{"service", "is_failed"},
	)

	UploadedFilesCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_upload_files_total",
			Help: "Total number of uploaded files to storage",
		},
		[]string{"service", "is_failed"},
	)

	CreatedProcessingTasksCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_created_tasks_total",
			Help: "Total number of created tasks of processing",
		},
		[]string{"service", "is_failed"},
	)

	ProcessingTasksStatusCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_processing_tasks_status_total",
			Help: "Total number of processing tasks with status",
		},
		[]string{"service", "status"},
	)

	PublishedTasksQueueCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_published_tasks_queue_total",
			Help: "Total number of published tasks to queue",
		},
		[]string{"service", "is_failed", "retries"},
	)

	OrchestratorProcessingCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "watchtower_orchestrator_processed_total",
			Help: "Total processed documents into orchestrator",
		},
		[]string{"service", "status"},
	)

	OrchestratorProcessingDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_orchestrator_processing_duration_seconds",
			Help: "Latency of full document processing time in seconds",
		},
		[]string{"service", "status"},
	)

	RecognizerDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_recognizer_duration_seconds",
			Help: "Latency of recognizing text from document file",
		},
		[]string{"service", "is_failed"},
	)

	StoreProcessedDocumentDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "watchtower_store_document_duration_seconds",
			Help: "Latency of storing processed document",
		},
		[]string{"service", "is_failed"},
	)
}
