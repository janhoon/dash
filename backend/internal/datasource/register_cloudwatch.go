package datasource

import (
	"time"

	acecw "github.com/aceobservability/ace-datasource-cloudwatch"

	"github.com/aceobservability/ace/backend/internal/models"
	"github.com/aceobservability/ace/backend/internal/ssrf"
)

func init() {
	register(models.DataSourceCloudWatch, func(ds models.DataSource) (*acecw.Client, error) {
		// Bare DatasourceClient only: dial/URL/redirect SSRF. Do not wrap
		// wrapDatasourceAuth — that stripper deletes Authorization on any
		// host that is not ds.URL. CloudWatch is dual-host (monitoring.* vs
		// logs.*) and signs with SDK SigV4, not Ace AuthType credentials.
		return acecw.New(configFromDataSource(ds), ssrf.DatasourceClient(30*time.Second))
	})
}
