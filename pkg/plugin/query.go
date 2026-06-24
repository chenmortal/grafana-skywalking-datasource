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

		// Handle "Query" query type
		if query.QueryType == "" {
			var frame *data.Frame
			queryTrace, err := dsInfo.SkywalkingClient.Trace(ctx, query.Query, q.TimeRange)

			frame = TransformTraceResponse(queryTrace, ctx, dsInfo, q)
			if err != nil {
				response.Responses[q.RefID] = backend.ErrorResponseWithErrorSource(err)
				continue
			}

			response.Responses[q.RefID] = backend.DataResponse{
				Frames: []*data.Frame{frame},
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

		traceName := fmt.Sprintf("%s %s", rootSpan.GetServiceCode(), *rootSpan.GetEndpointName())

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

func TransformTraceResponse(tracesResponse *queryV2TracesResponse, ctx context.Context, dsInfo *datasourceInfo, q backend.DataQuery) *data.Frame {
	frame := data.NewFrame(q.RefID,
		data.NewField("traceID", nil, []string{}),
		data.NewField("spanID", nil, []string{}),
		data.NewField("parentSpanID", nil, []*string{}),
		data.NewField("operationName", nil, []string{}),
		data.NewField("serviceName", nil, []string{}),
		data.NewField("serviceTags", nil, []json.RawMessage{}),
		data.NewField("startTime", nil, []int64{}),
		data.NewField("duration", nil, []int64{}),
		data.NewField("references", nil, []json.RawMessage{}),
	)
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: "trace",
		Custom: map[string]interface{}{
			"traceFormat": "skywalking",
		},
	}
	traces := tracesResponse.QueryTraces.GetTraces()
	if len(traces) == 0 {
		return frame
	}
	spans := traces[0].GetSpans()
	instanceTagsCache := map[string][]KeyValueType{}
	for _, span := range spans {
		spanID := transformSpanID(span.GetSegmentId(), span.GetSpanId())

		//  parse parentSpanId
		var parentSpanID *string
		if isRootSpan(span.GetParentSpanId()) {
			for _, ref := range span.GetRefs() {
				if ref.GetTraceId() == span.GetTraceId() && ref.GetParentSegmentId() != "" && ref.ParentSpanId >= 0 {
					s := transformSpanID(ref.GetParentSegmentId(), ref.GetParentSpanId())
					parentSpanID = &s
					break
				}
			}
		} else {
			s := transformSpanID(span.GetSegmentId(), span.GetParentSpanId())
			parentSpanID = &s
		}

		// parse operationName
		operationName := stringPtrValue(span.GetEndpointName())

		// parse serviceName
		serviceName := span.GetServiceCode()

		// parse layer (保持指针类型，用于 QueryInstancesByName)
		layer := span.GetLayer()

		// parse serviceTag
		var serviceTagList []KeyValueType
		instanceName := span.GetServiceInstanceName()
		if tagList, ok := instanceTagsCache[instanceName]; ok {
			serviceTagList = tagList
		} else {
			if queryResp, err := dsInfo.SkywalkingClient.QueryInstancesByName(ctx, serviceName, layer, q.TimeRange.From, q.TimeRange.To); err == nil {
				for _, pod := range queryResp.GetPods() {
					var serviceListTmp = []KeyValueType{
						{
							Key:   "instance",
							Type:  "string",
							Value: instanceName,
						},
						{
							Key:   "language",
							Type:  "string",
							Value: pod.GetLanguage(),
						},
						{
							Key:   "instanceUUID",
							Type:  "string",
							Value: pod.GetInstanceUUID(),
						},
					}
					for _, attr := range pod.Attributes {
						serviceListTmp = append(serviceListTmp, KeyValueType{
							Key:   attr.GetName(),
							Type:  "string",
							Value: attr.GetValue(),
						})
					}
					instanceTagsCache[pod.GetValue()] = serviceListTmp
					if pod.GetValue() == instanceName {
						serviceTagList = serviceListTmp
					}
				}
			}
		}

		serviceTags := json.RawMessage{}
		serviceTagMarshal, err := json.Marshal(serviceTagList)
		if err == nil {
			serviceTags = json.RawMessage(serviceTagMarshal)
		}

		// parse startTime
		startTime := span.GetStartTime()

		// parse duration
		duration := span.GetEndTime() - span.GetStartTime()

		// parse references
		references := json.RawMessage{}
		var refs = span.GetRefs()
		var traceReferences []TraceSpanReference
		for _, ref := range refs {
			transformSpanID := transformSpanID(ref.GetParentSegmentId(), ref.GetParentSpanId())
			if isRootSpan(ref.GetParentSpanId()) || transformSpanID != *parentSpanID {
				traceReferences = append(traceReferences, TraceSpanReference{
					RefType: convertRefType(string(ref.GetType())),
					TraceID: ref.GetTraceId(),
					SpanID:  transformSpanID,
				})
			}

		}
		refsMarshaled, err := json.Marshal(traceReferences)
		if err == nil {
			references = json.RawMessage(refsMarshaled)
		}

		frame.AppendRow(
			span.GetTraceId(),
			spanID,
			parentSpanID,
			operationName,
			serviceName,
			serviceTags,
			startTime,
			duration,
			references,
		)
	}
	return frame
}
func transformSpanID(segmentId string, spanId int) string {
	return fmt.Sprintf("%s-%d", segmentId, spanId)
}
func isRootSpan(spanId int) bool {
	return spanId == -1
}

type KeyValueType struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}
type TraceSpanReference struct {
	// RefType is not supported for OTLP-based traces and may be empty.
	RefType SpanRefType `json:"refType"`
	SpanID  string      `json:"spanID"`
	TraceID string      `json:"traceID"`
}

type SpanRefType string

const (
	SpanRefTypeChildOf     SpanRefType = "CHILD_OF"
	SpanRefTypeFollowsFrom SpanRefType = "FOLLOWS_FROM"
	SpanRefTypeExternal    SpanRefType = "EXTERNAL"
)

func convertRefType(refType string) SpanRefType {
	switch refType {
	case "CROSS_PROCESS":
		return SpanRefTypeChildOf
	case "CROSS_THREAD":
		return SpanRefTypeFollowsFrom
	default:
		return SpanRefTypeExternal
	}
}

// stringPtrValue 解引用字符串指针，nil 返回空字符串
func stringPtrValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
