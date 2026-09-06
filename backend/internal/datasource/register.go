package datasource

import (
	"github.com/aceobservability/ace/backend/internal/models"
	dscontract "github.com/aceobservability/ace/backend/pkg/datasource"
)

func init() {
	register(models.DataSourcePrometheus, NewPrometheusClient)
	register(models.DataSourceVictoriaMetrics, NewVictoriaMetricsClient)
	register(models.DataSourceLoki, NewLokiClient)
	register(models.DataSourceVictoriaLogs, NewVictoriaLogsClient)
	register(models.DataSourceTempo, NewTempoClient)
	register(models.DataSourceVictoriaTraces, NewVictoriaTracesClient)
	register(models.DataSourceClickHouse, NewClickHouseClient)
	register(models.DataSourceCloudWatch, NewCloudWatchClient)
	register(models.DataSourceElasticsearch, NewElasticsearchClient)
}

func register[T Client](typ models.DataSourceType, ctor func(models.DataSource) (T, error)) {
	dscontract.RegisterDatasource(string(typ), func(cfg dscontract.Config) (dscontract.Client, error) {
		return ctor(dataSourceFromConfig(cfg))
	})
}

func dataSourceFromConfig(cfg dscontract.Config) models.DataSource {
	return models.DataSource{
		Name:         cfg.Name,
		Type:         models.DataSourceType(cfg.Type),
		URL:          cfg.URL,
		AuthType:     cfg.AuthType,
		AuthConfig:   cfg.AuthConfig,
		TraceIDField: cfg.TraceIDField,
	}
}

func configFromDataSource(ds models.DataSource) dscontract.Config {
	return dscontract.Config{
		Name:         ds.Name,
		Type:         string(ds.Type),
		URL:          ds.URL,
		AuthType:     ds.AuthType,
		AuthConfig:   ds.AuthConfig,
		TraceIDField: ds.TraceIDField,
	}
}
