package plugin

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type SkywalkingClient struct {
	logger        log.Logger
	graphqlClient graphql.Client
}

func New(hc *http.Client, logger log.Logger, endpoint string) (SkywalkingClient, error) {
	//  settings.URL
	graphqlClient := graphql.NewClient(endpoint, hc)
	return SkywalkingClient{
		logger:        logger,
		graphqlClient: graphqlClient,
	}, nil
}

var VIRTUAL_LAYER = []string{"UNDEFINED", "VIRTUAL_DATABASE", "VIRTUAL_MQ", "VIRTUAL_GATEWAY"}

func (s SkywalkingClient) QueryV2Traces(ctx context.Context, condition TraceQueryCondition, timeRange backend.TimeRange) (*queryV2TracesResponse, error) {
	duration := convertToDuration(timeRange.From, timeRange.To)
	condition.QueryDuration = &duration
	resp, err := queryV2Traces(ctx, s.graphqlClient, &condition)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s SkywalkingClient) QueryBasicTraces(ctx context.Context, condition TraceQueryCondition, timeRange backend.TimeRange) (*queryTracesResponse, error) {
	duration := convertToDuration(timeRange.From, timeRange.To)
	condition.QueryDuration = &duration
	resp, err := queryTraces(ctx, s.graphqlClient, &condition)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s SkywalkingClient) TraceV2(ctx context.Context, traceId string, timeRange backend.TimeRange) (*queryV2TracesResponse, error) {
	duration := convertToDuration(timeRange.From, timeRange.To)
	condition := TraceQueryCondition{
		TraceId:       &traceId,
		TraceState:    TraceStateAll,
		QueryDuration: &duration,
		QueryOrder:    QueryOrderByStartTime,
	}
	resp, err := queryV2Traces(ctx, s.graphqlClient, &condition)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (s SkywalkingClient) TraceV1(ctx context.Context, traceId string, timeRange backend.TimeRange) (*querySpansResponse, error) {
	resp, err := querySpans(ctx, s.graphqlClient, traceId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s SkywalkingClient) ListLayer(ctx context.Context) (*listLayerResponse, error) {
	resp, err := listLayer(ctx, s.graphqlClient)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (s SkywalkingClient) QueryHasQueryTracesV2Support(ctx context.Context) (bool, error) {
	resp, err := queryHasQueryTracesV2Support(ctx, s.graphqlClient)
	if err != nil {
		return false, err
	}
	return resp.GetHasQueryTracesV2Support(), nil

}

func (s SkywalkingClient) QueryServices(ctx context.Context, layer string) (*queryServicesResponse, error) {
	resp, err := queryServices(ctx, s.graphqlClient, layer)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s SkywalkingClient) QueryEndpoints(ctx context.Context, serviceId, keyword string, fromTime, toTime time.Time, limit int) (*queryEndpointsResponse, error) {
	duration := convertToDuration(fromTime, toTime)
	resp, err := queryEndpoints(ctx, s.graphqlClient, serviceId, keyword, &duration, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s SkywalkingClient) QueryInstances(ctx context.Context, serviceId string, fromTime, toTime time.Time) (*queryInstancesResponse, error) {
	duration := convertToDuration(fromTime, toTime)
	resp, err := queryInstances(ctx, s.graphqlClient, serviceId, duration)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s SkywalkingClient) QueryInstancesByName(ctx context.Context, serviceName string, layer *string, fromTime, toTime time.Time) (*queryInstancesByNameResponse, error) {
	service := ServiceCondition{
		ServiceName: serviceName,
	}
	if layer != nil && slices.Contains(VIRTUAL_LAYER, *layer) {
		service.Layer = layer
	}
	duration := convertToDuration(fromTime, toTime)
	resp, err := queryInstancesByName(ctx, s.graphqlClient, service, duration)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func convertToDuration(fromTime, toTime time.Time) Duration {
	step := calculateStep(fromTime, toTime)
	duration := Duration{
		Start: formatTimeToString(fromTime, step),
		End:   formatTimeToString(toTime, step),
		Step:  step,
	}
	return duration
}

func formatTimeToString(t time.Time, step Step) string {
	switch step {
	case StepDay:
		return t.Format("2006-01-02")
	case StepHour:
		return t.Format("2006-01-02 15")
	case StepMinute:
		return t.Format("2006-01-02 1504")
	default:
		return t.Format("2006-01-02 150405")
	}
}

func calculateStep(from, to time.Time) Step {
	duration := to.Sub(from)
	switch {
	case duration <= time.Hour:
		return StepMinute
	case duration <= 24*time.Hour:
		return StepHour
	default:
		return StepDay
	}
}
