package server

import (
	"context"
	"time"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// GetPerformanceMetrics aggregates internal Prometheus metrics into a structured gRPC response.
func (s *GRPCService) GetPerformanceMetrics(ctx context.Context, _ *cdcpb.GetPerformanceMetricsRequest) (*cdcpb.GetPerformanceMetricsResponse, error) {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return nil, err
	}

	resp := &cdcpb.GetPerformanceMetricsResponse{
		Sinks:   make(map[string]*cdcpb.SinkPerformance),
		Sources: make(map[string]*cdcpb.SourcePerformance),
	}

	var producedCount float64
	var errorCount float64

	for _, mf := range mfs {
		switch mf.GetName() {
		case "cdc_events_produced_total":
			producedCount = s.sumCounter(mf)
			// Extract per-source stats
			for _, m := range mf.GetMetric() {
				instanceID := s.getLabelValue(m, "instance_id")
				status := s.getLabelValue(m, "status")
				if instanceID == "" {
					continue
				}
				if _, ok := resp.Sources[instanceID]; !ok {
					resp.Sources[instanceID] = &cdcpb.SourcePerformance{SourceId: instanceID}
				}
				val := m.GetCounter().GetValue()
				if status == "success" {
					resp.Sources[instanceID].Throughput = val / time.Since(s.startTime).Seconds()
				}
			}
		case "cdc_source_errors_total":
			for _, m := range mf.GetMetric() {
				instanceID := s.getLabelValue(m, "instance_id")
				if instanceID != "" && resp.Sources[instanceID] != nil {
					resp.Sources[instanceID].ErrorRate += m.GetCounter().GetValue()
				}
			}
		case "cdc_sink_errors_total":
			errorCount = s.sumCounter(mf)
		case "cdc_active_workers":
			resp.ActiveWorkers = uint32(s.sumGauge(mf))
		case "cdc_sink_write_duration_seconds":
			resp.LatencyP99 = s.getP99FromHistogram(mf)
			// Extract per-sink stats
			for _, m := range mf.GetMetric() {
				sinkID := s.getLabelValue(m, "sink_id")
				if sinkID == "" {
					continue
				}
				if _, ok := resp.Sinks[sinkID]; !ok {
					resp.Sinks[sinkID] = &cdcpb.SinkPerformance{SinkId: sinkID}
				}
				// Rough avg latency from histogram sum/count
				if m.GetHistogram().GetSampleCount() > 0 {
					resp.Sinks[sinkID].AvgLatency = m.GetHistogram().GetSampleSum() / float64(m.GetHistogram().GetSampleCount()) * 1000
				}
				// Add throughput for sinks too
				if m.GetHistogram().GetSampleCount() > 0 {
					resp.Sinks[sinkID].Throughput = float64(m.GetHistogram().GetSampleCount()) / time.Since(s.startTime).Seconds()
				}
			}
		}
	}

	uptime := time.Since(s.startTime).Seconds()
	if uptime > 0 {
		resp.Throughput = producedCount / uptime
		if producedCount > 0 {
			resp.ErrorRate = (errorCount / producedCount) * 100
		}
	}

	return resp, nil
}

func (s *GRPCService) sumCounter(mf *dto.MetricFamily) float64 {
	var total float64
	for _, m := range mf.GetMetric() {
		total += m.GetCounter().GetValue()
	}
	return total
}

func (s *GRPCService) sumGauge(mf *dto.MetricFamily) float64 {
	var total float64
	for _, m := range mf.GetMetric() {
		total += m.GetGauge().GetValue()
	}
	return total
}

func (s *GRPCService) getLabelValue(m *dto.Metric, name string) string {
	for _, l := range m.GetLabel() {
		if l.GetName() == name {
			return l.GetValue()
		}
	}
	return ""
}

func (s *GRPCService) getP99FromHistogram(mf *dto.MetricFamily) float64 {
	// Simple P99 approximation: find the first bucket that contains 99% of samples
	for _, m := range mf.GetMetric() {
		h := m.GetHistogram()
		if h.GetSampleCount() == 0 {
			continue
		}
		target := 0.99 * float64(h.GetSampleCount())
		for _, b := range h.GetBucket() {
			if float64(b.GetCumulativeCount()) >= target {
				return b.GetUpperBound() * 1000 // Convert to ms
			}
		}
	}
	return 0
}
