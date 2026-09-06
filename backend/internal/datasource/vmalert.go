package datasource

import (
	"time"

	acevmalert "github.com/aceobservability/ace-datasource-vmalert"

	"github.com/aceobservability/ace/backend/internal/models"
)

func NewVMAlertClient(ds models.DataSource) (*acevmalert.Client, error) {
	return acevmalert.New(ds.URL, newDatasourceHTTPClient(ds, 30*time.Second))
}
