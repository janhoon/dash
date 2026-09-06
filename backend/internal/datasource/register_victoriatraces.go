package datasource

import (
	"time"

	acevt "github.com/aceobservability/ace-datasource-victoriatraces"

	"github.com/aceobservability/ace/backend/internal/models"
)

func init() {
	register(models.DataSourceVictoriaTraces, func(ds models.DataSource) (*acevt.Client, error) {
		return acevt.New(ds.URL, newDatasourceHTTPClient(ds, 15*time.Second))
	})
}
