package datasource

import (
	"time"

	acevl "github.com/aceobservability/ace-datasource-victorialogs"

	"github.com/aceobservability/ace/backend/internal/models"
)

func init() {
	register(models.DataSourceVictoriaLogs, func(ds models.DataSource) (*acevl.Client, error) {
		return acevl.New(ds.URL, newDatasourceHTTPClient(ds, 30*time.Second))
	})
}
