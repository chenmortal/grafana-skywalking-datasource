package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
)

var logger = backend.NewLoggerWith("logger", "skywalking")

type Service struct {
	im instancemgmt.InstanceManager
}

func ProvideService(httpClientProvider *httpclient.Provider) *Service {
	return &Service{
		im: datasource.NewInstanceManager(newInstanceSettings(httpClientProvider)),
	}
}

type datasourceInfo struct {
	SkywalkingClient SkywalkingClient
	SourceSettings   backend.DataSourceInstanceSettings
	PluginSettings   PluginSettings
}

type PluginSettings struct {
	Path string `json:"path"`
	V2   bool   `json:"interfacev2"`
}

func newInstanceSettings(httpClientProvider *httpclient.Provider) datasource.InstanceFactoryFunc {
	return func(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
		httpClientOptions, err := settings.HTTPClientOptions(ctx)
		if err != nil {
			return nil, backend.DownstreamError(fmt.Errorf("error reading settings: %w", err))
		}

		httpClient, err := httpClientProvider.New(httpClientOptions)
		if err != nil {
			return nil, fmt.Errorf("error creating http client: %w", err)
		}

		if settings.URL == "" {
			return nil, backend.DownstreamError(errors.New("error reading settings: url is empty"))
		}

		logger := logger.FromContext(ctx)
		skywalkingClient, err := New(httpClient, logger, settings.URL)
		if err != nil {
			return nil, fmt.Errorf("error creating skywalking client: %w", err)
		}

		pluginSettings := PluginSettings{
			V2: false,
		}
		if err := json.Unmarshal(settings.JSONData, &pluginSettings); err != nil {
			return nil, fmt.Errorf("error parsing plugin settings: %w", err)
		}

		return &datasourceInfo{SkywalkingClient: skywalkingClient, SourceSettings: settings, PluginSettings: pluginSettings}, err
	}
}
func (s *Service) getDSInfo(ctx context.Context, pluginCtx backend.PluginContext) (*datasourceInfo, error) {
	i, err := s.im.Get(ctx, pluginCtx)
	if err != nil {
		return nil, err
	}

	instance, ok := i.(*datasourceInfo)
	if !ok {
		return nil, errors.New("failed to cast datasource info")
	}

	return instance, nil
}
func (s *Service) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	client, err := s.getDSInfo(ctx, backend.PluginConfigFromContext(ctx))
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}
	if client.PluginSettings.V2 {
		support, error := client.SkywalkingClient.QueryHasQueryTracesV2Support(ctx)
		if error != nil {
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: error.Error(),
			}, nil
		}
		if !support {
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "Data source doesn't support queryTracesV2, please set false",
			}, nil
		}
	}
	_, err = client.SkywalkingClient.ListLayer(ctx)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

func (s *Service) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	handler := httpadapter.New(s.registerResourceRoutes())
	return handler.CallResource(ctx, req, sender)
}

func (s *Service) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	dsInfo, err := s.getDSInfo(ctx, req.PluginContext)
	if err != nil {
		return nil, err
	}
	return queryData(ctx, dsInfo, req)
}
