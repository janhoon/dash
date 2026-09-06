package datasource

import (
	"time"

	acech "github.com/aceobservability/ace-datasource-clickhouse"

	"github.com/aceobservability/ace/backend/internal/models"
)

func init() {
	register(models.DataSourceClickHouse, func(ds models.DataSource) (*acech.Client, error) {
		return acech.New(ds.URL, newDatasourceHTTPClient(ds, 30*time.Second), ds.AuthConfig)
	})
}
