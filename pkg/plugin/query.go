package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type SkywalkingQuery struct {
	Layer     string              `json:"layer"`
	Query     string              `json:"query"`
	QueryType string              `json:"queryType"`
	Condition TraceQueryCondition `json:"condition"`
}

func queryData(ctx context.Context, dsInfo *datasourceInfo, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		var query SkywalkingQuery

		err := json.Unmarshal(q.JSON, &query)
		if err != nil {
			err = backend.DownstreamError(fmt.Errorf("error while parsing the query json. %w", err))
			response.Responses[q.RefID] = backend.ErrorResponseWithErrorSource(err)
			continue
		}

		// Handle "Search" query type
		if query.QueryType == "search" {
			queryTrace, err := dsInfo.SkywalkingClient.Search(ctx, query.Condition, q.TimeRange)
			frames := TransformSearchResponse(queryTrace, dsInfo.Settings.UID, dsInfo.Settings.Name)
			if err != nil {
				response.Responses[q.RefID] = backend.ErrorResponseWithErrorSource(err)
				continue
			}

			response.Responses[q.RefID] = backend.DataResponse{
				Frames: data.Frames{frames},
			}
		}

	}

	return response, nil
}

func TransformSearchResponse(tracesResponse *queryV2TracesResponse, dsUID string, dsName string) *data.Frame {
	frame := data.NewFrame("traces",
		data.NewField("traceID", nil, []string{}).SetConfig(&data.FieldConfig{
			DisplayName: "Trace ID",
			Links: []data.DataLink{
				{
					Title: "Trace: ${__value.raw}",
					URL:   "",
					Internal: &data.InternalDataLink{
						DatasourceUID:  dsUID,
						DatasourceName: dsName,
						Query: map[string]interface{}{
							"query": "${__value.raw}",
						},
					},
				},
			},
		}),
		data.NewField("traceName", nil, []string{}).SetConfig(&data.FieldConfig{
			DisplayName: "Trace name",
		}),
		data.NewField("startTime", nil, []time.Time{}).SetConfig(&data.FieldConfig{
			DisplayName: "Start time",
			Unit:        "ms",
		}),
		data.NewField("duration", nil, []int64{}).SetConfig(&data.FieldConfig{
			DisplayName: "Duration",
			Unit:        "ms",
		}),
	)
	// Set the visualization type to table
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: "table",
	}
	for _, trace := range tracesResponse.QueryTraces.Traces {
		spans := trace.GetSpans()
		if len(spans) == 0 {
			continue
		}
		rootSpan := spans[0]
		for _, span := range spans {
			if span.GetStartTime() < rootSpan.GetStartTime() {
				rootSpan = span
			}
		}

		traceName := fmt.Sprintf("%s %s", rootSpan.GetServiceCode(), rootSpan.GetEndpointName())

		startTime := time.UnixMilli(rootSpan.GetStartTime())

		frame.AppendRow(
			rootSpan.GetTraceId(),
			traceName,
			startTime,
			rootSpan.GetEndTime()-rootSpan.GetStartTime(),
		)
	}
	return frame
}
