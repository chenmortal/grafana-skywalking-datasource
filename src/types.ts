import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';
import { TraceQueryCondition } from 'type/operations';

export interface SkywalkingQuery extends DataQuery {
  layer?: string;
  query: string;
  condition: TraceQueryCondition;
  endpointQueryText?: string;
  queryType?: SkywalkingQueryType;
}

export const DEFAULT_QUERY: Partial<SkywalkingQuery> = {
  layer: 'GENERAL',
  condition: {
    paging: {
      pageSize: 20,
    },
    queryOrder: 'BY_START_TIME',
    traceState: 'ALL',
  },
};

export interface DataPoint {
  Time: number;
  Value: number;
}

export interface DataSourceResponse {
  datapoints: DataPoint[];
}

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  apiKey?: string;
}

export type SkywalkingQueryType = 'search' | 'dependencyGraph';
