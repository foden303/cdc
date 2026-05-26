package ports

import "context"

type MetricsReader interface {
	FlowProcessingLatencyP99(ctx context.Context, window string) (float64, error)
}
