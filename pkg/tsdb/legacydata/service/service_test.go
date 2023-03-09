package service

import (
	"context"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/plugins"
	acmock "github.com/grafana/grafana/pkg/services/accesscontrol/mock"
	"github.com/grafana/grafana/pkg/services/datasources"
	datasourceservice "github.com/grafana/grafana/pkg/services/datasources/service"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/quota/quotatest"
	"github.com/grafana/grafana/pkg/services/secrets/fakes"
	secretskvs "github.com/grafana/grafana/pkg/services/secrets/kvstore"
	secretsmng "github.com/grafana/grafana/pkg/services/secrets/manager"
	"github.com/grafana/grafana/pkg/tsdb/legacydata"
)

func TestHandleRequest(t *testing.T) {
	t.Run("Should invoke plugin manager QueryData when handling request for query", func(t *testing.T) {
		client := &fakePluginsClient{}
		var actualReq *backend.QueryDataRequest
		client.QueryDataHandlerFunc = func(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
			actualReq = req
			return backend.NewQueryDataResponse(), nil
		}
		sqlStore := db.InitTestDB(t)
		secretsService := secretsmng.SetupTestService(t, fakes.NewFakeSecretsStore())
		secretsStore := secretskvs.NewSQLSecretsKVStore(sqlStore, secretsService, log.New("test.logger"))
		datasourcePermissions := acmock.NewMockedPermissionsService()
		quotaService := quotatest.New(false, nil)
		dsService, err := datasourceservice.ProvideService(nil, secretsService, secretsStore, sqlStore.Cfg, featuremgmt.WithFeatures(), acmock.New(), datasourcePermissions, quotaService)
		require.NoError(t, err)

		s := ProvideService(client, nil, dsService)

		ds := &datasources.DataSource{ID: 12, Type: "unregisteredType", JsonData: simplejson.New()}
		req := legacydata.DataQuery{
			TimeRange: &legacydata.DataTimeRange{},
			Queries: []legacydata.DataSubQuery{
				{RefID: "A", DataSource: &datasources.DataSource{ID: 1, Type: "test"}, Model: simplejson.New()},
				{RefID: "B", DataSource: &datasources.DataSource{ID: 1, Type: "test"}, Model: simplejson.New()},
			},
		}
		res, err := s.HandleRequest(context.Background(), ds, req)
		require.NoError(t, err)
		require.NotNil(t, actualReq)
		require.NotNil(t, res)
	})
}

type fakePluginsClient struct {
	plugins.Client
	backend.QueryDataHandlerFunc
}

func (m *fakePluginsClient) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	if m.QueryDataHandlerFunc != nil {
		return m.QueryDataHandlerFunc.QueryData(ctx, req)
	}

	return nil, nil
}
