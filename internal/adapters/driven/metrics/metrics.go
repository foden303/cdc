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

	SourceErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdc_source_errors_total",
		Help: "Total number of non-terminal errors in source capture",
	}, []string{"instance_id", "error_type"})

	// Sink metrics
	EventsConsumedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdc_events_consumed_total",
		Help: "Total number of events consumed from WAL and sent to sink",
	}, []string{"instance_id", "status"}) // status: success/failure

	SinkErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdc_sink_errors_total",
		Help: "Total number of non-terminal errors in sink delivery",
	}, []string{"instance_id", "sink_id", "error_type"})

	// Flow metrics
	DLQEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdc_dlq_events_total",
		Help: "Total number of events moved to the dead-letter-queue",
	}, []string{"flow_id", "reason"})

	ActiveWorkers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cdc_active_workers",
		Help: "Number of active parallel partition workers in engine",
	}, []string{"type"}) // type: producer/worker

	// Performance metrics
	WorkerProcessDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cdc_worker_process_duration_seconds",
		Help:    "Time spent processing an event in a worker",
		Buckets: prometheus.DefBuckets,
	}, []string{"sink_id"})

	SinkWriteDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cdc_sink_write_duration_seconds",
		Help:    "Time spent writing to a sink",
		Buckets: prometheus.DefBuckets,
	}, []string{"sink_id", "type"})

	// NATS connection health
	NATSReconnectTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cdc_nats_reconnect_total",
		Help: "Total number of NATS reconnections",
	})

	// Backpressure monitoring
	EventChannelUtilization = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cdc_event_channel_utilization",
		Help: "Ratio of event channel usage (0.0 to 1.0), high values indicate backpressure",
	})

	// Batch size distribution
	BatchSizeHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cdc_batch_size",
		Help:    "Distribution of actual batch sizes being processed",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"component"}) // component: producer/worker

	// Per-flow metrics
	FlowEventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdc_flow_events_processed_total",
		Help: "Total number of events processed by a flow",
	}, []string{"flow_id", "status"}) // status: success/failure

	FlowReplicationLag = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cdc_flow_replication_lag_ms",
		Help: "Replication lag in milliseconds per flow",
	}, []string{"flow_id"})

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
