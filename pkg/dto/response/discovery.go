package response

import "github.com/foden/cdc/pkg/interfaces"

type DiscoverTablesResponse struct {
	Tables []interfaces.TableInfo
}

type DiscoverSinkTablesResponse struct {
	Tables []interfaces.TableInfo
}
