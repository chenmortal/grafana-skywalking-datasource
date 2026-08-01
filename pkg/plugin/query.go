package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
			var frames *data.Frame
			if dsInfo.PluginSettings.V2 {
				queryTrace, err := dsInfo.SkywalkingClient.QueryV2Traces(ctx, query.Condition, q.TimeRange)
				frames = TransformQueryV2Response(queryTrace, dsInfo.SourceSettings.UID, dsInfo.SourceSettings.Name)
				if err != nil {
					response.Responses[q.RefID] = backend.ErrorResponseWithErrorSource(err)
					continue
				}
			} else {
				queryTrace, err := dsInfo.SkywalkingClient.QueryBasicTraces(ctx, query.Condition, q.TimeRange)
				frames = TransformQueryBasicResponse(queryTrace, dsInfo.SourceSettings.UID, dsInfo.SourceSettings.Name)
				if err != nil {
					response.Responses[q.RefID] = backend.ErrorResponseWithErrorSource(err)
					continue
				}
			}

			response.Responses[q.RefID] = backend.DataResponse{
				Frames: data.Frames{frames},
			}
		}

		// Handle "Query" query type
		if query.QueryType == "" {
			if query.Query == "" {
				response.Responses[q.RefID] = backend.DataResponse{}
				continue
			}

			var frame *data.Frame
			if dsInfo.PluginSettings.V2 {
				queryTrace, err := dsInfo.SkywalkingClient.TraceV2(ctx, query.Query, q.TimeRange)
				frame = TransformTraceV2Response(queryTrace, ctx, dsInfo, q)
				if err != nil {
					response.Responses[q.RefID] = backend.ErrorResponseWithErrorSource(err)
					continue
				}
			} else {
				queryTrace, err := dsInfo.SkywalkingClient.TraceV1(ctx, query.Query, q.TimeRange)
				frame = TransformTraceV1Response(queryTrace, ctx, dsInfo, q)
				if err != nil {
					response.Responses[q.RefID] = backend.ErrorResponseWithErrorSource(err)
					continue
				}
			}

			response.Responses[q.RefID] = backend.DataResponse{
				Frames: []*data.Frame{frame},
			}
		}

	}

	return response, nil
}

func getSearchFrame(dsUID string, dsName string) *data.Frame {
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
		data.NewField("endpointName", nil, []string{}).SetConfig(&data.FieldConfig{
			DisplayName: "Endpoint name",
		}),
		data.NewField("startTime", nil, []time.Time{}).SetConfig(&data.FieldConfig{
			DisplayName: "Start time",
			Unit:        "ms",
		}),
		data.NewField("duration", nil, []int64{}).SetConfig(&data.FieldConfig{
			DisplayName: "Duration",
			Unit:        "ms",
		}),
		data.NewField("isError", nil, []bool{}),
	)
	// Set the visualization type to table
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: "table",
	}
	return frame
}

func TransformQueryV2Response(tracesResponse *queryV2TracesResponse, dsUID string, dsName string) *data.Frame {
	frame := getSearchFrame(dsUID, dsName)
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

		endpointName := *rootSpan.GetEndpointName()

		startTime := time.UnixMilli(rootSpan.GetStartTime())

		frame.AppendRow(
			rootSpan.GetTraceId(),
			endpointName,
			startTime,
			rootSpan.GetEndTime()-rootSpan.GetStartTime(),
			*rootSpan.IsError,
		)
	}
	return frame
}
func TransformQueryBasicResponse(tracesResponse *queryTracesResponse, dsUID string, dsName string) *data.Frame {
	frame := getSearchFrame(dsUID, dsName)
	for _, trace := range tracesResponse.GetData().GetTraces() {
		startTimeMillis, err := strconv.ParseInt(trace.GetStart(), 10, 64)
		if err != nil {
			// 如果解析失败，使用当前时间作为默认值
			startTimeMillis = time.Now().UnixMilli()
		}
		startTime := time.UnixMilli(startTimeMillis)

		frame.AppendRow(
			trace.GetTraceIds()[0],
			trace.GetEndpointNames()[0],
			startTime,
			int64(trace.Duration),
			*trace.IsError,
		)
	}
	return frame
}

type spanAccessor interface {
	GetTraceId() string
	GetSegmentId() string
	GetSpanId() int
	GetParentSpanId() int
	GetEndpointName() *string
	GetServiceCode() string
	GetServiceInstanceName() string
	GetStartTime() int64
	GetEndTime() int64
	GetPeer() *string
	GetLayer() *string
	GetComponent() *string
}

type refAccessor interface {
	GetTraceId() string
	GetParentSegmentId() string
	GetParentSpanId() int
	GetType() RefType
}

type tagAccessor interface {
	GetKey() string
	GetValue() *string
}

func newTraceFrame(refID string) *data.Frame {
	frame := data.NewFrame(refID,
		data.NewField("traceID", nil, []string{}),
		data.NewField("spanID", nil, []string{}),
		data.NewField("parentSpanID", nil, []*string{}),
		data.NewField("operationName", nil, []string{}),
		data.NewField("serviceName", nil, []string{}),
		data.NewField("serviceTags", nil, []json.RawMessage{}),
		data.NewField("startTime", nil, []int64{}),
		data.NewField("duration", nil, []int64{}),
		data.NewField("references", nil, []json.RawMessage{}),
		data.NewField("tags", nil, []json.RawMessage{}),
	)
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: "trace",
		Custom: map[string]interface{}{
			"traceFormat": "skywalking",
		},
	}
	return frame
}

func transformTraceResponse(
	refID string,
	spans []spanAccessor,
	refsList [][]refAccessor,
	tagsList [][]tagAccessor,
	ctx context.Context,
	dsInfo *datasourceInfo,
	q backend.DataQuery,
) *data.Frame {
	frame := newTraceFrame(refID)
	instanceTagsCache := map[string][]KeyValueType{}
	for i, span := range spans {
		spanID := transformSpanID(span.GetSegmentId(), span.GetSpanId())

		var parentSpanID *string
		if isRootSpan(span.GetParentSpanId()) {
			for _, ref := range refsList[i] {
				if ref.GetTraceId() == span.GetTraceId() && ref.GetParentSegmentId() != "" && ref.GetParentSpanId() >= 0 {
					s := transformSpanID(ref.GetParentSegmentId(), ref.GetParentSpanId())
					parentSpanID = &s
					break
				}
			}
		} else {
			s := transformSpanID(span.GetSegmentId(), span.GetParentSpanId())
			parentSpanID = &s
		}

		references := buildReferences(refsList[i], parentSpanID)
		tags := buildTags(span.GetPeer(), span.GetLayer(), span.GetComponent(), tagsList[i])
		serviceTags := buildServiceTags(ctx, dsInfo, instanceTagsCache, q, span.GetServiceCode(), span.GetLayer(), span.GetServiceInstanceName())

		frame.AppendRow(
			span.GetTraceId(), spanID, parentSpanID, stringPtrValue(span.GetEndpointName()),
			span.GetServiceCode(), serviceTags, span.GetStartTime(),
			span.GetEndTime()-span.GetStartTime(), references, tags,
		)
	}
	return frame
}

func TransformTraceV2Response(tracesResponse *queryV2TracesResponse, ctx context.Context, dsInfo *datasourceInfo, q backend.DataQuery) *data.Frame {
	traces := tracesResponse.QueryTraces.GetTraces()
	if len(traces) == 0 {
		return newTraceFrame(q.RefID)
	}
	spans := traces[0].GetSpans()
	spanAccessors := make([]spanAccessor, len(spans))
	refsList := make([][]refAccessor, len(spans))
	tagsList := make([][]tagAccessor, len(spans))
	for i := range spans {
		spanAccessors[i] = &spans[i]
		refs := spans[i].GetRefs()
		refsList[i] = make([]refAccessor, len(refs))
		for j := range refs {
			refsList[i][j] = &refs[j]
		}
		tags := spans[i].GetTags()
		tagsList[i] = make([]tagAccessor, len(tags))
		for j := range tags {
			tagsList[i][j] = &tags[j]
		}
	}
	return transformTraceResponse(q.RefID, spanAccessors, refsList, tagsList, ctx, dsInfo, q)
}

func TransformTraceV1Response(tracesResponse *querySpansResponse, ctx context.Context, dsInfo *datasourceInfo, q backend.DataQuery) *data.Frame {
	spans := tracesResponse.GetTrace().GetSpans()
	spanAccessors := make([]spanAccessor, len(spans))
	refsList := make([][]refAccessor, len(spans))
	tagsList := make([][]tagAccessor, len(spans))
	for i := range spans {
		spanAccessors[i] = &spans[i]
		refs := spans[i].GetRefs()
		refsList[i] = make([]refAccessor, len(refs))
		for j := range refs {
			refsList[i][j] = &refs[j]
		}
		tags := spans[i].GetTags()
		tagsList[i] = make([]tagAccessor, len(tags))
		for j := range tags {
			tagsList[i][j] = &tags[j]
		}
	}
	return transformTraceResponse(q.RefID, spanAccessors, refsList, tagsList, ctx, dsInfo, q)
}
func transformSpanID(segmentId string, spanId int) string {
	return fmt.Sprintf("%s-%d", segmentId, spanId)
}
func isRootSpan(spanId int) bool {
	return spanId == -1
}

func buildReferences(refs []refAccessor, parentSpanID *string) json.RawMessage {
	var traceReferences []TraceSpanReference
	for _, ref := range refs {
		transformedSpanID := transformSpanID(ref.GetParentSegmentId(), ref.GetParentSpanId())
		if isRootSpan(ref.GetParentSpanId()) || transformedSpanID != *parentSpanID {
			traceReferences = append(traceReferences, TraceSpanReference{
				RefType: convertRefType(string(ref.GetType())),
				TraceID: ref.GetTraceId(),
				SpanID:  transformedSpanID,
			})
		}
	}
	references := json.RawMessage{}
	refsMarshaled, err := json.Marshal(traceReferences)
	if err == nil {
		references = json.RawMessage(refsMarshaled)
	}
	return references
}

func buildTags(peer *string, layer *string, component *string, tags []tagAccessor) json.RawMessage {
	var tagList = []KeyValueType{}
	if peer != nil {
		tagList = append(tagList, KeyValueType{Key: "peer", Type: "string", Value: *peer})
	}
	if layer != nil {
		tagList = append(tagList, KeyValueType{Key: "layer", Type: "string", Value: *layer})
	}
	if component != nil {
		tagList = append(tagList, KeyValueType{Key: "component", Type: "string", Value: *component})
	}
	for _, tag := range tags {
		tagList = append(tagList, KeyValueType{Key: tag.GetKey(), Type: "string", Value: tag.GetValue()})
	}
	result := json.RawMessage{}
	tagMarshal, err := json.Marshal(tagList)
	if err == nil {
		result = json.RawMessage(tagMarshal)
	}
	return result
}

func buildServiceTags(ctx context.Context, dsInfo *datasourceInfo,
	instanceTagsCache map[string][]KeyValueType, q backend.DataQuery,
	serviceName string, layer *string, instanceName string) json.RawMessage {
	var serviceTagList []KeyValueType
	if tagList, ok := instanceTagsCache[instanceName]; ok {
		serviceTagList = tagList
	} else {
		if queryResp, err := dsInfo.SkywalkingClient.QueryInstancesByName(ctx, serviceName, layer, q.TimeRange.From, q.TimeRange.To); err == nil {
			for _, pod := range queryResp.GetPods() {
				var serviceListTmp = []KeyValueType{
					{Key: "instance", Type: "string", Value: instanceName},
					{Key: "language", Type: "string", Value: pod.GetLanguage()},
					{Key: "instanceUUID", Type: "string", Value: pod.GetInstanceUUID()},
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
	return serviceTags
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
