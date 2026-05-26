package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Source metrics
	EventsProducedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdc_events_produced_total",
		Help: "Total number of events captured from source and sent to WAL",
	}, []string{"instance_id", "status"}) // status: success/failure

	// Flow metrics
	DLQEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdc_dlq_events_total",
		Help: "Total number of events moved to the dead-letter-queue",
	}, []string{"flow_id", "reason"})

	// Performance metrics
	SinkWriteDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cdc_sink_write_duration_seconds",
		Help:    "Time spent writing to a sink",
		Buckets: prometheus.DefBuckets,
	}, []string{"sink_id", "type"})

	FlowProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cdc_flow_processing_duration_seconds",
		Help:    "Time from worker batch processing start to successful sink write for a flow",
		Buckets: prometheus.DefBuckets,
	}, []string{"flow_id"})

	// NATS connection health
	NATSReconnectTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cdc_nats_reconnect_total",
		Help: "Total number of NATS reconnections",
	})

	// Per-flow metrics
	FlowEventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdc_flow_events_processed_total",
		Help: "Total number of events processed by a flow",
	}, []string{"flow_id", "status"}) // status: success/failure

	FlowWorkerPoolActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cdc_flow_worker_pool_active",
		Help: "Number of active workers in a flow's worker pool",
	}, []string{"flow_id"})

	FlowWorkerPoolCapacity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cdc_flow_worker_pool_capacity",
		Help: "Total capacity of a flow's worker pool",
	}, []string{"flow_id"})

	FlowBatchSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cdc_flow_batch_size",
		Help:    "Distribution of batch sizes processed per flow",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"flow_id"})
)
