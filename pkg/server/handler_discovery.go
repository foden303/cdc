package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/dto/request"
	"github.com/foden/cdc/pkg/interfaces"
)

func (s *CDCService) DiscoverTables(ctx context.Context, req *cdcpb.DiscoverTablesRequest) (*cdcpb.DiscoverTablesResponse, error) {
	result, err := s.discoveryService.DiscoverSourceTables(ctx, request.DiscoverTablesRequest{SourceID: req.SourceId})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to discover tables: %v", err)
	}

	return &cdcpb.DiscoverTablesResponse{Tables: tablesToProto(result.Tables)}, nil
}

func (s *CDCService) DiscoverSinkTables(ctx context.Context, req *cdcpb.DiscoverSinkTablesRequest) (*cdcpb.DiscoverSinkTablesResponse, error) {
	result, err := s.discoveryService.DiscoverSinkTables(ctx, request.DiscoverSinkTablesRequest{SinkID: req.SinkId})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to discover sink tables: %v", err)
	}

	return &cdcpb.DiscoverSinkTablesResponse{Tables: tablesToProto(result.Tables)}, nil
}

func tablesToProto(tables []interfaces.TableInfo) []*cdcpb.TableInfo {
	var pbTables []*cdcpb.TableInfo
	for _, t := range tables {
		pbTable := &cdcpb.TableInfo{
			Schema: t.Schema,
			Name:   t.Name,
		}
		for _, c := range t.Columns {
			pbTable.Columns = append(pbTable.Columns, &cdcpb.ColumnInfo{
				Name:         c.Name,
				Type:         c.Type,
				IsPrimaryKey: c.IsPrimaryKey,
				IsNullable:   c.IsNullable,
			})
		}
		pbTables = append(pbTables, pbTable)
	}
	return pbTables
}
