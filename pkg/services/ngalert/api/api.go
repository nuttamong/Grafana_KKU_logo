package api

import (
	"time"

	"github.com/grafana/grafana/pkg/services/quota"

	"github.com/grafana/grafana/pkg/services/ngalert/metrics"
	"github.com/grafana/grafana/pkg/services/ngalert/state"

	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/datasourceproxy"
	"github.com/grafana/grafana/pkg/services/datasources"
	apimodels "github.com/grafana/grafana/pkg/services/ngalert/api/tooling/definitions"
	"github.com/grafana/grafana/pkg/services/ngalert/schedule"
	"github.com/grafana/grafana/pkg/services/ngalert/store"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/tsdb"
)

// timeNow makes it possible to test usage of time
var timeNow = time.Now

type Alertmanager interface {
	// Configuration
	SaveAndApplyConfig(config *apimodels.PostableUserConfig) error

	// Silences
	CreateSilence(ps *apimodels.PostableSilence) (string, error)
	DeleteSilence(silenceID string) error
	GetSilence(silenceID string) (apimodels.GettableSilence, error)
	ListSilences(filter []string) (apimodels.GettableSilences, error)

	// Alerts
	GetAlerts(active, silenced, inhibited bool, filter []string, receiver string) (apimodels.GettableAlerts, error)
	GetAlertGroups(active, silenced, inhibited bool, filter []string, receiver string) (apimodels.AlertGroups, error)
}

// API handlers.
type API struct {
	Cfg             *setting.Cfg
	DatasourceCache datasources.CacheService
	RouteRegister   routing.RouteRegister
	DataService     *tsdb.Service
	QuotaService    *quota.QuotaService
	Schedule        schedule.ScheduleService
	RuleStore       store.RuleStore
	InstanceStore   store.InstanceStore
	AlertingStore   store.AlertingStore
	DataProxy       *datasourceproxy.DatasourceProxyService
	Alertmanager    Alertmanager
	StateManager    *state.Manager
}

// RegisterAPIEndpoints registers API handlers
func (api *API) RegisterAPIEndpoints(m *metrics.Metrics) {
	logger := log.New("ngalert.api")
	proxy := &AlertingProxy{
		DataProxy: api.DataProxy,
	}

	// Register endpoints for proxing to Alertmanager-compatible backends.
	api.RegisterAlertmanagerApiEndpoints(NewForkedAM(
		api.DatasourceCache,
		NewLotexAM(proxy, logger),
		AlertmanagerSrv{store: api.AlertingStore, am: api.Alertmanager, log: logger},
	), m)
	// Register endpoints for proxing to Prometheus-compatible backends.
	api.RegisterPrometheusApiEndpoints(NewForkedProm(
		api.DatasourceCache,
		NewLotexProm(proxy, logger),
		PrometheusSrv{log: logger, manager: api.StateManager, store: api.RuleStore},
	), m)
	// Register endpoints for proxing to Cortex Ruler-compatible backends.
	api.RegisterRulerApiEndpoints(NewForkedRuler(
		api.DatasourceCache,
		NewLotexRuler(proxy, logger),
		RulerSrv{DatasourceCache: api.DatasourceCache, QuotaService: api.QuotaService, manager: api.StateManager, store: api.RuleStore, log: logger},
	), m)
	api.RegisterTestingApiEndpoints(TestingApiSrv{
		AlertingProxy:   proxy,
		Cfg:             api.Cfg,
		DataService:     api.DataService,
		DatasourceCache: api.DatasourceCache,
		log:             logger,
	}, m)
}
