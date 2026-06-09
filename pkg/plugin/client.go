package plugin

import (
	"context"
	"net/http"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type SkywalkingClient struct {
	logger        log.Logger
	graphqlClient graphql.Client
	settings      backend.DataSourceInstanceSettings
}

func New(hc *http.Client, logger log.Logger, settings backend.DataSourceInstanceSettings) (SkywalkingClient, error) {
	//  settings.URL
	graphqlClient := graphql.NewClient(settings.URL, hc)
	return SkywalkingClient{
		logger:        logger,
		graphqlClient: graphqlClient,
		settings:      settings,
	}, nil
}

func (s SkywalkingClient) ListLayer(ctx context.Context) (*listLayerResponse, error) {
	resp, err := listLayer(ctx, s.graphqlClient)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (s SkywalkingClient) QueryServices(ctx context.Context, layer string) (*queryServicesResponse, error) {
	resp, err := queryServices(ctx, s.graphqlClient, layer)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (s SkywalkingClient) QueryEndpoints(ctx context.Context, serviceId, keyword string, fromTime, toTime time.Time, limit int) (*queryEndpointsResponse, error) {
	step := calculateStep(fromTime, toTime)
	duration := Duration{
		Start:     formatTimeToString(fromTime, step),
		End:       formatTimeToString(toTime, step),
		Step:      step,
		ColdStage: false,
	}
	resp, err := queryEndpoints(ctx, s.graphqlClient, serviceId, keyword, duration, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
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
