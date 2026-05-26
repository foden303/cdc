package infrastructure

import (
	"context"
	"fmt"

	"github.com/foden/cdc/config"
	drivennats "github.com/foden/cdc/internal/adapters/driven/nats"
)

func SetupNATS(ctx context.Context, cfg config.NATSConfig) (*drivennats.Client, error) {
	natsClient, err := drivennats.NewClient(&cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}
	if err := natsClient.CreateStream(ctx, []string{"cdc.>"}); err != nil {
		return nil, fmt.Errorf("create CDC stream: %w", err)
	}
	if err := natsClient.CreateDLQStream(ctx); err != nil {
		return nil, fmt.Errorf("create DLQ stream: %w", err)
	}
	return natsClient, nil
}
