import {
  DataSourceInstanceSettings,
  CoreApp,
  ScopedVars,
  DataQueryRequest,
  DataQueryResponse,
  FieldType,
  toDataFrame,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { SkywalkingQuery, MyDataSourceOptions, DEFAULT_QUERY } from './types';
import { ListLayerQuery, QueryEndpointsQuery, QueryInstancesQuery, QueryServicesQuery } from 'type/operations';
import { getTimeRangeValues } from 'utils';
import { map, Observable, of } from 'rxjs';

export class SkywalkingDataSource extends DataSourceWithBackend<SkywalkingQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }
  query(options: DataQueryRequest<SkywalkingQuery>): Observable<DataQueryResponse> {
    const target: SkywalkingQuery = options.targets[0];
    console.log('target', target);
    if (!target) {
      return of({ data: [emptyTraceDataFrame] });
    }
    return super.query({ ...options, targets: [target] }).pipe(
      map((response) => {
        console.log('response', response);
        return response;
      })
    );
  }
  getDefaultQuery(_: CoreApp): Partial<SkywalkingQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: SkywalkingQuery, scopedVars: ScopedVars) {
    return {
      ...query,
      queryText: getTemplateSrv().replace(query.endpointQueryText, scopedVars),
    };
  }

  async listLayers() {
    const response = await this.getResource<ListLayerQuery>('layers');
    return response.layers || [DEFAULT_QUERY.layer];
  }

  async queryServices(layer: string) {
    const response = await this.getResource<QueryServicesQuery>(
      `services/${encodeURIComponent(getTemplateSrv().replace(layer!))}`
    );
    return response.services || [];
  }

  async queryEndpoints(serviceId: string | number, keyword: string, limit: number) {
    const range = getTimeRangeValues();
    const data = {
      serviceId: serviceId,
      keyword: keyword,
      fromTime: range.from,
      toTime: range.to,
      limit: limit,
    };
    const response = await this.postResource<QueryEndpointsQuery>('endpoints', data);
    return response.pods || [];
  }

  async queryInstances(serviceId: string | number) {
    const range = getTimeRangeValues();
    const data = {
      serviceId: serviceId,
      fromTime: range.from,
      toTime: range.to,
    };
    const response = await this.postResource<QueryInstancesQuery>('instances', data);
    return response.pods || [];
  }

  // filterQuery(query: SkywalkingQuery): boolean {
  //   return !!query.queryText;
  // }
}

const emptyTraceDataFrame = toDataFrame({
  fields: [{ name: 'trace', type: FieldType.trace, values: [] }],
  meta: {
    preferredVisualisationType: 'trace',
    custom: {
      traceFormat: 'skywalking',
    },
  },
});
