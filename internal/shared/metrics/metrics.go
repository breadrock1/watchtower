package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	StoredArticleCounter            prometheus.Counter
	StoredMatchedArticleTagsCounter prometheus.Counter
	CreatedMonitoringTasksCounter   prometheus.Counter
)

func init() {
	StoredArticleCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "monitoring_stored_articles_total",
		Help: "Total number of stored articles",
	})

	StoredMatchedArticleTagsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "monitoring_stored_matched_articles_total",
		Help: "Total number of stored matched article tags",
	})

	CreatedMonitoringTasksCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "monitoring_created_tasks_total",
		Help: "Total number of created tasks",
	})
}
