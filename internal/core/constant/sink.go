package constant

type SinkType string

const (
	SinkTypePostgres      SinkType = "postgres"
	SinkTypeClickhouse    SinkType = "clickhouse"
	SinkTypeElasticsearch SinkType = "elasticsearch"
)

func (s SinkType) String() string {
	return string(s)
}
