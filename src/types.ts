import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface SkywalkingQuery extends DataQuery {
  layer: string;
  query: string | null | undefined;
  queryText?: string;
  constant: number;
  queryType?: SkywalkingQueryType;
  serviceId?: string;
  endpoint?: string;
}

export const DEFAULT_QUERY: Partial<SkywalkingQuery> = {
  constant: 6.5,
  layer: 'GENERAL',
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


