import { DataQueryRequest, DataSourceInstanceSettings, MutableDataFrame } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { firstValueFrom, of } from 'rxjs';
import { SkywalkingDataSource } from './datasource';
import { ALL_OPERATIONS_VALUE, SkywalkingDataSourceOptions, SkywalkingQuery } from './types';

const instanceSettings = {
  url: '',
  name: '',
  type: 'grafana-skywalking-datasource',
  uid: 'skywalking',
  meta: {},
} as unknown as DataSourceInstanceSettings<SkywalkingDataSourceOptions>;

function buildRequest(targets: SkywalkingQuery[]): DataQueryRequest<SkywalkingQuery> {
  return {
    targets,
    requestId: 'test',
    interval: '',
    intervalMs: 0,
    range: {} as DataQueryRequest['range'],
    scopedVars: {},
    timezone: '',
    app: 'panel-editor',
    startTime: 0,
    endTime: 0,
  } as unknown as DataQueryRequest<SkywalkingQuery>;
}

function searchTarget(refId: string, overrides: Partial<SkywalkingQuery> = {}): SkywalkingQuery {
  return {
    refId,
    query: '',
    queryType: 'search',
    condition: {
      serviceId: `service-${refId}`,
      endpointId: '',
      queryOrder: 'BY_START_TIME',
      traceState: 'ALL',
      paging: { pageSize: 20 },
    },
    ...overrides,
  } as SkywalkingQuery;
}

describe('SkywalkingDataSource.query', () => {
  let superQuerySpy: jest.SpyInstance;

  beforeEach(() => {
    superQuerySpy = jest
      .spyOn(DataSourceWithBackend.prototype, 'query')
      .mockReturnValue(of({ data: [] as MutableDataFrame[] }));
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  it('passes all targets to the backend, preserving order and refIds', async () => {
    const ds = new SkywalkingDataSource(instanceSettings);
    const targets = [searchTarget('A'), searchTarget('B'), searchTarget('C')];

    await firstValueFrom(ds.query(buildRequest(targets)));

    expect(superQuerySpy).toHaveBeenCalledTimes(1);
    const request = superQuerySpy.mock.calls[0][0] as DataQueryRequest<SkywalkingQuery>;
    expect(request.targets).toHaveLength(3);
    expect(request.targets.map((t) => t.refId)).toEqual(['A', 'B', 'C']);
  });

  it('sanitizes the condition of every target', async () => {
    const ds = new SkywalkingDataSource(instanceSettings);
    const targets = [
      searchTarget('A', {
        condition: {
          serviceId: '',
          endpointId: ALL_OPERATIONS_VALUE,
          queryOrder: 'BY_START_TIME',
        },
      } as Partial<SkywalkingQuery>),
      searchTarget('B'),
    ];

    await firstValueFrom(ds.query(buildRequest(targets)));

    const request = superQuerySpy.mock.calls[0][0] as DataQueryRequest<SkywalkingQuery>;
    expect(request.targets[0].condition.serviceId).toBeNull();
    expect(request.targets[0].condition.endpointId).toBeNull();
    expect(request.targets[0].condition.queryOrder).toBe('BY_START_TIME');
    // second target keeps its own condition untouched where valid
    expect(request.targets[1].condition.serviceId).toBe('service-B');
    expect(request.targets[1].condition.endpointId).toBeNull();
  });

  it('returns an empty trace frame without calling the backend when there are no targets', async () => {
    const ds = new SkywalkingDataSource(instanceSettings);

    const response = await firstValueFrom(ds.query(buildRequest([])));

    expect(superQuerySpy).not.toHaveBeenCalled();
    expect(response.data).toHaveLength(1);
    expect(response.data[0].meta?.preferredVisualisationType).toBe('trace');
  });

  it('keeps hidden targets so the backend still processes them', async () => {
    const ds = new SkywalkingDataSource(instanceSettings);
    const targets = [searchTarget('A'), searchTarget('B', { hide: true })];

    await firstValueFrom(ds.query(buildRequest(targets)));

    const request = superQuerySpy.mock.calls[0][0] as DataQueryRequest<SkywalkingQuery>;
    expect(request.targets).toHaveLength(2);
    expect(request.targets[1].hide).toBe(true);
  });

  it('leaves targets without a condition unchanged', async () => {
    const ds = new SkywalkingDataSource(instanceSettings);
    const traceIdTarget = { refId: 'A', query: 'abc123', condition: undefined } as unknown as SkywalkingQuery;

    await firstValueFrom(ds.query(buildRequest([traceIdTarget])));

    const request = superQuerySpy.mock.calls[0][0] as DataQueryRequest<SkywalkingQuery>;
    expect(request.targets[0].condition).toBeUndefined();
  });
});
