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

export type Step =
  | 'DAY'
  | 'HOUR'
  | 'MINUTE'
  | 'SECOND';

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
