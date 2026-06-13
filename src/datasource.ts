import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { SkywalkingQuery, MyDataSourceOptions, DEFAULT_QUERY } from './types';
import { ListLayerQuery, QueryEndpointsQuery, QueryInstancesQuery, QueryServicesQuery } from 'type/operations';
import { getTimeRangeValues } from 'utils';


export class SkywalkingDataSource extends DataSourceWithBackend<SkywalkingQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<SkywalkingQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: SkywalkingQuery, scopedVars: ScopedVars) {
    return {
      ...query,
      queryText: getTemplateSrv().replace(query.queryText, scopedVars),
    };
  }

  async listLayers(){
    const response=await this.getResource<ListLayerQuery>('layers');
    return response.layers || [DEFAULT_QUERY.layer];
  }

  async queryServices(layer: string){
    const response= await this.getResource<QueryServicesQuery>(`services/${encodeURIComponent(getTemplateSrv().replace(layer!))}`);
    return response.services || []
  }

  async queryEndpoints(serviceId:string,keyword:string,limit:number){
    const range=getTimeRangeValues()
    const data={
      serviceId: serviceId,
      keyword:keyword,
      fromTime:range.from,
      toTime:range.to,
      limit:limit,
    };
    const response= await this.postResource<QueryEndpointsQuery>("endpoints",data);
    return response.pods || []
  }

  async queryInstances(serviceId:string){
    const range=getTimeRangeValues()
    const data={
      serviceId: serviceId,
      fromTime:range.from,
      toTime:range.to,
    };
    const response= await this.postResource<QueryInstancesQuery>("instances",data);
    return response.pods || []
  }

  filterQuery(query: SkywalkingQuery): boolean {
    // if no query has been provided, prevent the query from being executed
    return !!query.queryText;
  }
}


