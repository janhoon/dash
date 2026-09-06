package datasource

import (
	"time"

	acetempo "github.com/aceobservability/ace-datasource-tempo"

	"github.com/aceobservability/ace/backend/internal/models"
)

func init() {
	register(models.DataSourceTempo, func(ds models.DataSource) (*acetempo.Client, error) {
		return acetempo.New(ds.URL, newDatasourceHTTPClient(ds, 15*time.Second))
	})
}
