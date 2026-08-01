/** Internal type. DO NOT USE DIRECTLY. */
type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
/** Internal type. DO NOT USE DIRECTLY. */
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
export type Duration = {
  coldStage?: boolean | null | undefined;
  end: string;
  start: string;
  step: Step;
};

export type Language =
  | 'DOTNET'
  | 'GO'
  | 'JAVA'
  | 'LUA'
  | 'NODEJS'
  | 'PHP'
  | 'PYTHON'
  | 'RUBY'
  | 'UNKNOWN';

export type Pagination = {
  pageNum?: number | null | undefined;
  pageSize: number;
};

export type QueryOrder =
  | 'BY_DURATION'
  | 'BY_START_TIME';

export type RefType =
  | 'CROSS_PROCESS'
  | 'CROSS_THREAD';

export type SpanTag = {
  key: string;
  value?: string | null | undefined;
};

export type Step =
  | 'DAY'
  | 'HOUR'
  | 'MINUTE'
  | 'SECOND';

export type TraceQueryCondition = {
  endpointId?: string | number | null | undefined;
  maxTraceDuration?: number | null | undefined;
  minTraceDuration?: number | null | undefined;
  paging: Pagination;
  queryDuration?: Duration | null | undefined;
  queryOrder: QueryOrder;
  serviceId?: string | number | null | undefined;
  serviceInstanceId?: string | number | null | undefined;
  tags?: Array<SpanTag> | null | undefined;
  traceId?: string | null | undefined;
  traceState: TraceState;
};

export type TraceState =
  | 'ALL'
  | 'ERROR'
  | 'SUCCESS';

export type QueryServicesQueryVariables = Exact<{
  layer: string;
}>;


export type QueryServicesQuery = { services: Array<{ id: string, group: string, layers: Array<string>, normal: boolean | null, shortName: string, value: string, label: string }> };

export type ListLayerQueryVariables = Exact<{ [key: string]: never; }>;


export type ListLayerQuery = { layers: Array<string> };

export type QueryEndpointsQueryVariables = Exact<{
  serviceId: string | number;
  keyword: string;
  duration?: Duration | null | undefined;
  limit: number;
}>;


export type QueryEndpointsQuery = { pods: Array<{ id: string, value: string, label: string }> };

export type QueryInstancesQueryVariables = Exact<{
  serviceId: string | number;
  duration: Duration;
}>;


export type QueryInstancesQuery = { pods: Array<{ id: string, language: Language, instanceUUID: string, value: string, label: string, attributes: Array<{ name: string, value: string }> }> };

export type QueryV2TracesQueryVariables = Exact<{
  condition?: TraceQueryCondition | null | undefined;
}>;


export type QueryV2TracesQuery = { queryTraces: { traces: Array<{ spans: Array<{ traceId: string, segmentId: string, spanId: number, parentSpanId: number, serviceCode: string, serviceInstanceName: string, startTime: number, endTime: number, endpointName: string | null, type: string, peer: string | null, component: string | null, isError: boolean | null, layer: string | null, refs: Array<{ traceId: string, parentSegmentId: string, parentSpanId: number, type: RefType }>, tags: Array<{ key: string, value: string | null }>, logs: Array<{ time: number, data: Array<{ key: string, value: string | null }> | null }>, attachedEvents: Array<{ event: string, startTime: { seconds: number, nanos: number }, endTime: { seconds: number, nanos: number }, tags: Array<{ key: string, value: string | null } | null>, summary: Array<{ key: string, value: number }> }> }> }>, retrievedTimeRange: { startTime: number, endTime: number } } | null };
