import { css } from '@emotion/css';
import { Combobox, ComboboxOption, InlineField, InlineFieldRow, RadioButtonGroup, Stack } from '@grafana/ui';
import { SkywalkingDataSource } from '../datasource';
import { ALL_OPERATIONS_VALUE, DEFAULT_QUERY, SkywalkingQuery } from '../types';
import { descodeEndpointName, descodeServiceID, useTimeRangeFromUrl } from 'utils';
import { QueryOrder, TraceState } from 'type/operations';
import { t } from '@grafana/i18n';
import React, { useState, useCallback, useEffect } from 'react';

type Props = {
  datasource: SkywalkingDataSource;
  query: SkywalkingQuery;
  onChange: (value: SkywalkingQuery) => void;
};

export const ALL_OPERATIONS_KEY = 'All';

export const ALL_COMBOBOX_OPTION: ComboboxOption<string> = {
  label: ALL_OPERATIONS_KEY,
  value: ALL_OPERATIONS_VALUE,
};

export function SearchForm({ datasource, query, onChange }: Props) {
  const [layerOptions, setLayerOptions] = useState<ComboboxOption[]>([]);
  const [serviceOptions, setServiceOptions] = useState<Array<ComboboxOption<string>>>([]);
  const [endpointOptions, setEndpointOptions] = useState<Array<ComboboxOption<string>>>([]);
  const [serviceInstanceIdOptions, setServiceInstanceIdOptions] = useState<Array<ComboboxOption<string>>>([]);
  const [selectedEndpoint, setSelectedEndpoint] = useState<ComboboxOption<string> | null>(null);
  const { from,to } = useTimeRangeFromUrl();
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const layerList = await datasource.listLayers();
        if (!cancelled) {
          const options = layerList.map((l: string) => ({ label: l, value: l }));
          setLayerOptions(options);
        }
      } catch (error) {
        console.error('Failed to load layers:', error);
        if (!cancelled) {
          setLayerOptions([{ label: DEFAULT_QUERY.layer, value: DEFAULT_QUERY.layer! }]);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [datasource]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const services = await datasource.queryServices(query.layer!);
        if (cancelled) {
          return;
        }
        setServiceOptions([
          ALL_COMBOBOX_OPTION,
          ...services.map((s: { label: string; id: string; group?: string }) => ({
            label: s.label,
            value: s.id,
            group: s.group,
          })),
        ]);
      } catch (error) {
        if (cancelled) {
          return;
        }
        console.error('Failed to load services:', error);
        setServiceOptions([ALL_COMBOBOX_OPTION]);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [datasource, query.layer]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      if (!query.condition.serviceId || query.condition.serviceId === ALL_COMBOBOX_OPTION.value) {
        setEndpointOptions([ALL_COMBOBOX_OPTION]);
        setServiceInstanceIdOptions([ALL_COMBOBOX_OPTION]);
        return;
      }
      try {
        const [endpointsResult, instancesResult] = await Promise.allSettled([
          datasource.queryEndpoints(query.condition.serviceId, '', 20),
          datasource.queryInstances(query.condition.serviceId),
        ]);
        if (cancelled) {
          return;
        }
        if (endpointsResult.status === 'fulfilled') {
          setEndpointOptions([
            ALL_COMBOBOX_OPTION,
            ...endpointsResult.value.map((e: { label: string; id: string }) => ({
              label: e.label,
              value: e.id,
              description: 'service : ' + descodeServiceID(e.id),
            })),
          ]);
        } else {
          console.error('Failed to fetch endpoint options:', endpointsResult.reason);
          setEndpointOptions([ALL_COMBOBOX_OPTION]);
        }
        if (instancesResult.status === 'fulfilled') {
          setServiceInstanceIdOptions([
            ALL_COMBOBOX_OPTION,
            ...instancesResult.value.map((i: { label: string; id: string }) => ({
              label: i.label,
              value: i.id,
            })),
          ]);
        } else {
          console.error('Failed to fetch instance options:', instancesResult.reason);
          setServiceInstanceIdOptions([ALL_COMBOBOX_OPTION]);
        }
      } catch (error) {
        if (cancelled) {
          return;
        }
        console.error('Failed to fetch endpoint options:', error);
        setEndpointOptions([ALL_COMBOBOX_OPTION]);
        setServiceInstanceIdOptions([ALL_COMBOBOX_OPTION]);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [datasource, query.condition.serviceId,from,to]);

  const fetchEndpointOptions = useCallback(
    async (keyword: string): Promise<Array<ComboboxOption<string>>> => {
      if (!query.condition.serviceId || query.condition.serviceId === ALL_COMBOBOX_OPTION.value) {
        return [ALL_COMBOBOX_OPTION];
      }
      if (!keyword.trim()) {
        return endpointOptions.length > 0 ? endpointOptions : [ALL_COMBOBOX_OPTION];
      }
      try {
        const endpoints = await datasource.queryEndpoints(query.condition.serviceId, keyword, 20);
        return [
          ALL_COMBOBOX_OPTION,
          ...endpoints.map((e) => ({
            label: e.label,
            value: e.id,
            description: 'service : ' + descodeServiceID(e.id),
          })),
        ];
      } catch (error) {
        console.error('Failed to fetch endpoint options:', error);
        return [ALL_COMBOBOX_OPTION];
      }
    },
    [datasource, query.condition.serviceId, endpointOptions]
  );

  // 辅助函数：更新 query.condition 中的指定字段
  const updateCondition = <K extends keyof SkywalkingQuery['condition']>(
    key: K,
    value: SkywalkingQuery['condition'][K],
    keysToReset: Array<keyof SkywalkingQuery['condition']> = []
  ) => {
    const condition = { ...query.condition, [key]: value };
    keysToReset.forEach((k) => {
      (condition as Record<string, unknown>)[k] = undefined;
    });
    onChange({
      ...query,
      condition,
    });
  };

  return (
    <>
      <div className={css({ maxWidth: '800px' })}>
        <Stack gap={1} alignItems="center" justifyContent="space-between">
          <div>
            <InlineFieldRow>
              <InlineField label={t('searchForm.traceState.label', 'state')} labelWidth={8}>
                <RadioButtonGroup<TraceState>
                  value={query.condition.traceState}
                  options={[
                    {
                      label: t('searchForm.traceState.options.all', 'ALL'),
                      value: 'ALL',
                    },
                    {
                      label: t('searchForm.traceState.options.success', 'SUCCESS'),
                      value: 'SUCCESS',
                    },
                    {
                      label: t('searchForm.traceState.options.error', 'ERROR'),
                      value: 'ERROR',
                    },
                  ]}
                  onChange={(v) => updateCondition('traceState', v!)}
                ></RadioButtonGroup>
              </InlineField>
              <InlineField label={t('searchForm.order.label', 'order')} labelWidth={8}>
                <RadioButtonGroup<QueryOrder>
                  value={query.condition.queryOrder}
                  options={[
                    { label: t('searchForm.order.options.slowest', 'slowest'), value: 'BY_DURATION' },
                    { label: t('searchForm.order.options.newest', 'newest'), value: 'BY_START_TIME' },
                  ]}
                  onChange={(v) => updateCondition('queryOrder', v!)}
                ></RadioButtonGroup>
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField label={t('searchForm.layer.label', 'Layer')} labelWidth={14} grow>
                <Combobox
                  options={layerOptions}
                  value={query.layer || 'GENERAL'}
                  onChange={(v) =>
                    onChange({
                      ...query,
                      layer: v?.value!,
                    })
                  }
                  placeholder="Select a layer"
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField label={t('searchForm.service.label', 'Service')} labelWidth={14} grow>
                <Combobox
                  options={serviceOptions}
                  value={query.condition.serviceId ?? null}
                  onChange={(v) => updateCondition('serviceId', v?.value)}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField label={t('searchForm.endpoint.label', 'Endpoint')} labelWidth={14} grow>
                <Combobox
                  options={fetchEndpointOptions}
                  value={
                    selectedEndpoint && selectedEndpoint.value === query.condition.endpointId
                      ? selectedEndpoint
                      : query.condition.endpointId
                        ? ({
                            label: descodeEndpointName(String(query.condition.endpointId)),
                            value: query.condition.endpointId,
                          } as ComboboxOption<string>)
                        : null
                  }
                  onChange={(v) => {
                    setSelectedEndpoint(v);
                    updateCondition('endpointId', v?.value, ['serviceInstanceId']);
                  }}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField label={t('searchForm.instance.label', 'Instance')} labelWidth={14} grow>
                <Combobox
                  options={serviceInstanceIdOptions}
                  value={
                    serviceInstanceIdOptions.find((v) => v?.value === query.condition.serviceInstanceId) || undefined
                  }
                  onChange={(v) => {
                    onChange({
                      ...query,
                      condition: {
                        ...query.condition,
                        serviceInstanceId: v?.value!,
                      },
                    });
                  }}
                />
              </InlineField>
            </InlineFieldRow>
          </div>
          <div></div>
        </Stack>
      </div>
    </>
  );
}

export default SearchForm;
