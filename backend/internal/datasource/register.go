package datasource

import (
	"time"

	aceprom "github.com/aceobservability/ace-datasource-prometheus"

	"github.com/aceobservability/ace/backend/internal/models"
	dscontract "github.com/aceobservability/ace/backend/pkg/datasource"
)

func init() {
	register(models.DataSourcePrometheus, func(ds models.DataSource) (*aceprom.Client, error) {
		return aceprom.New(ds.URL, newDatasourceHTTPClient(ds, 30*time.Second))
	})
	register(models.DataSourceLoki, NewLokiClient)
	register(models.DataSourceVictoriaLogs, NewVictoriaLogsClient)
	register(models.DataSourceTempo, NewTempoClient)
	register(models.DataSourceVictoriaTraces, NewVictoriaTracesClient)
	register(models.DataSourceCloudWatch, NewCloudWatchClient)
	register(models.DataSourceElasticsearch, NewElasticsearchClient)
}

func register[T Client](typ models.DataSourceType, ctor func(models.DataSource) (T, error)) {
	dscontract.RegisterDatasource(string(typ), func(cfg dscontract.Config) (dscontract.Client, error) {
		return ctor(dataSourceFromConfig(cfg))
	})
}

// dataSourceFromConfig and configFromDataSource map the module Config, not the
// stored row. ID, OrganizationID, IsDefault, LinkedTraceDatasourceID, CreatedAt,
// and UpdatedAt are omitted on purpose. After a registry round-trip, constructors
// must not rely on those fields.
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
