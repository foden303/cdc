package infrastructure

import (
	"context"
	"fmt"

	drivennats "github.com/foden/cdc/internal/adapters/driven/nats"
	drivenstorage "github.com/foden/cdc/internal/adapters/driven/storage"
	"github.com/foden/cdc/internal/core/ports"
)

func SetupStorage(ctx context.Context, natsClient *drivennats.Client) (ports.Store, error) {
	store, err := drivenstorage.NewNATSKVStore(ctx, natsClient.JetStream())
	if err != nil {
		return nil, fmt.Errorf("create storage: %w", err)
	}
	return store, nil
}
