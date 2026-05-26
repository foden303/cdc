package response

import "github.com/foden/cdc/internal/core/ports"

type DiscoverTablesResponse struct {
	Tables []ports.TableInfo
}

type DiscoverSinkTablesResponse struct {
	Tables []ports.TableInfo
}
