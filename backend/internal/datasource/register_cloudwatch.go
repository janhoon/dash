package datasource

import (
	"time"

	acecw "github.com/aceobservability/ace-datasource-cloudwatch"

	"github.com/aceobservability/ace/backend/internal/models"
)

func init() {
	register(models.DataSourceCloudWatch, func(ds models.DataSource) (*acecw.Client, error) {
		return acecw.New(configFromDataSource(ds), newDatasourceHTTPClient(ds, 30*time.Second))
	})
}
