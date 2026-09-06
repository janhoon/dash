package datasource

import (
	"time"

	acevm "github.com/aceobservability/ace-datasource-victoriametrics"

	"github.com/aceobservability/ace/backend/internal/models"
)

func init() {
	register(models.DataSourceVictoriaMetrics, func(ds models.DataSource) (*acevm.Client, error) {
		return acevm.New(ds.URL, newDatasourceHTTPClient(ds, 30*time.Second))
	})
}
