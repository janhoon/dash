package datasource

import (
	"time"

	acealert "github.com/aceobservability/ace-datasource-alertmanager"

	"github.com/aceobservability/ace/backend/internal/models"
)

func NewAlertManagerClient(ds models.DataSource) (*acealert.Client, error) {
	return acealert.New(ds.URL, newDatasourceHTTPClient(ds, 30*time.Second))
}
