import { css } from '@emotion/css';
import { useCallback, useEffect, useState } from 'react';
import { Combobox, ComboboxOption, InlineField, InlineFieldRow } from '@grafana/ui';

import { SkywalkingDataSource } from '../datasource';
import { SkywalkingQuery } from '../types';
import React from 'react';
import { descodeServiceID, useTimeRangeFromUrl } from 'utils';
type Props = {
  datasource: SkywalkingDataSource;
  query: SkywalkingQuery;
  onChange: (value: SkywalkingQuery) => void;
};

export const ALL_OPERATIONS_KEY = 'All';

const DEFAULT_LAYER = { label: 'GENERAL', value: 'GENERAL' };
export function SearchForm({ datasource, query, onChange }: Props) {
  const [layerOptions, setLayerOptions] = useState<ComboboxOption[]>([DEFAULT_LAYER]);
  const [serviceOptions, setServiceOptions] = useState<ComboboxOption<string>[]>([]);
  const [endpointOptions, setEndpointOptions] = useState<ComboboxOption<string>[]>([]);

  const { from, to } = useTimeRangeFromUrl();

  useEffect(() => {
    loadEndpoints();
  }, [from, to]);

  const loadLayers = useCallback(async () => {
    try {
      const layerList = await datasource.listLayers();
      const options = layerList.map((l: string) => ({ label: l, value: l }));
      setLayerOptions(options);
    } catch (error) {
      console.error('Failed to load layers:', error);
      setLayerOptions([DEFAULT_LAYER]);
    }
  }, [datasource]);

  const loadServices = useCallback(async () => {
    try {
      const services = await datasource.queryServices(query.layer);
      const serviceOptions: ComboboxOption<string>[] = services.map((s: any) => ({
        label: s.label,
        value: s.id,
        group: s.group,
      }));
      setServiceOptions(serviceOptions);
    } catch (error) {
      console.error('Failed to load services:', error);
      setServiceOptions([]);
    }
  }, [datasource, query.layer]);

  const loadEndpoints = useCallback(async () => {
    try {
      const endpoints = await datasource.queryEndpoints(query.serviceId!, query.queryText!, 20);
      setEndpointOptions(
        endpoints.map((e) => ({ label: e.label, value: e.id, description: 'service : ' + descodeServiceID(e.id) }))
      );
    } catch (error) {
      console.error('Failed to load endpoints:', error);
      setEndpointOptions([]);
    }
  }, [datasource, query.serviceId, query.queryText]);

  useEffect(() => {
    loadServices();
  }, [datasource, query.layer]);

  useEffect(() => {
    onChange({
      ...query,
      endpoint: undefined,
    });

    if (query.serviceId) {
      loadEndpoints();
    } else {
      setEndpointOptions([]);
    }
  }, [query.serviceId, loadEndpoints]);

  useEffect(() => {
    onChange({
      ...query,
      serviceId: undefined,
    });

    if (query.layer) {
      loadServices();
    } else {
      setServiceOptions([]);
    }
    loadLayers();
  }, [query.layer, loadServices]);

  return (
    <>
      <div className={css({ maxWidth: '500px' })}>
        <InlineFieldRow>
          <InlineField label="Layer" labelWidth={14} grow>
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
          <InlineField label="Service" labelWidth={14} grow>
            <Combobox
              options={serviceOptions}
              value={serviceOptions.find((v) => v?.value === query.serviceId) || undefined}
              onChange={(v) => {
                onChange({
                  ...query,
                  serviceId: v?.value!,
                });
              }}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="Endpoint" labelWidth={14} grow>
            <Combobox
              options={endpointOptions}
              value={endpointOptions.find((v) => v?.value === query.endpoint) || undefined}
              onChange={(v) => {
                onChange({
                  ...query,
                  endpoint: v?.value!,
                });
              }}
            />
          </InlineField>
        </InlineFieldRow>

        {/* <InlineFieldRow>
          <InlineField label="Operation Name" labelWidth={14} grow disabled={!query.service}>
            <Select
              inputId="operation"
              options={operationOptions}
              onOpenMenu={() =>
                loadOptions(
                  `services/${encodeURIComponent(getTemplateSrv().replace(query.service!))}/operations`,
                  'operations'
                )
              }
              isLoading={isLoading.operations}
              value={operationOptions?.find((v) => v.value === query.operation) || null}
              placeholder="Select an operation"
              onChange={(v) =>
                onChange({
                  ...query,
                  operation: v?.value! || undefined,
                })
              }
              menuPlacement="bottom"
              isClearable
              aria-label={'select-operation-name'}
              allowCustomValue={true}
            />
          </InlineField>
        </InlineFieldRow> */}
        {/* <InlineFieldRow>
          <InlineField label="Tags" labelWidth={14} grow tooltip="Values should be in logfmt.">
            <Input
              id="tags"
              value={query.tags}
              placeholder="http.status_code=200 error=true"
              onChange={(v) =>
                onChange({
                  ...query,
                  tags: v.currentTarget.value,
                })
              }
            />
          </InlineField>
        </InlineFieldRow> */}
        {/* <InlineFieldRow>
          <InlineField label="Min Duration" labelWidth={14} grow>
            <Input
              id="minDuration"
              name="minDuration"
              value={query.minDuration || ''}
              placeholder={durationPlaceholder}
              onChange={(v) =>
                onChange({
                  ...query,
                  minDuration: v.currentTarget.value,
                })
              }
            />
          </InlineField>
        </InlineFieldRow> */}
        {/* <InlineFieldRow>
          <InlineField label="Max Duration" labelWidth={14} grow>
            <Input
              id="maxDuration"
              name="maxDuration"
              value={query.maxDuration || ''}
              placeholder={durationPlaceholder}
              onChange={(v) =>
                onChange({
                  ...query,
                  maxDuration: v.currentTarget.value,
                })
              }
            />
          </InlineField>
        </InlineFieldRow> */}
        {/* <InlineFieldRow>
          <InlineField label="Limit" labelWidth={14} grow tooltip="Maximum number of returned results">
            <Input
              id="limit"
              name="limit"
              value={query.limit || ''}
              type="number"
              onChange={(v) =>
                onChange({
                  ...query,
                  limit: v.currentTarget.value ? parseInt(v.currentTarget.value, 10) : undefined,
                })
              }
            />
          </InlineField>
        </InlineFieldRow> */}
      </div>
    </>
  );
}

export default SearchForm;
