package datasource

import (
	"time"

	aceloki "github.com/aceobservability/ace-datasource-loki"

	"github.com/aceobservability/ace/backend/internal/models"
)

func init() {
	register(models.DataSourceLoki, func(ds models.DataSource) (*aceloki.Client, error) {
		return aceloki.New(ds.URL, newDatasourceHTTPClient(ds, 30*time.Second))
	})
}
