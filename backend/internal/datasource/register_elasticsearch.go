package datasource

import (
	"time"

	acees "github.com/aceobservability/ace-datasource-elasticsearch"

	"github.com/aceobservability/ace/backend/internal/models"
)

func init() {
	register(models.DataSourceElasticsearch, func(ds models.DataSource) (*acees.Client, error) {
		return acees.New(ds.URL, ds.AuthConfig, newDatasourceHTTPClient(ds, 30*time.Second))
	})
}
