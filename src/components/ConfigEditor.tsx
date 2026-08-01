import React from 'react';
import { Divider, InlineField, InlineFieldRow, InlineSwitch } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { SkywalkingDataSourceOptions, MySecureJsonData } from '../types';
import { ConnectionSettings } from '@grafana/plugin-ui';
interface Props extends DataSourcePluginOptionsEditorProps<SkywalkingDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;

  return (
    <>
      <Divider spacing={4} />
      <ConnectionSettings
        config={options}
        onChange={onOptionsChange}
        urlPlaceholder="http://skywalking.example.com/graphql"
      />
      <InlineFieldRow>
        <InlineField label="Interface Version  v2" labelWidth={24}>
          <InlineSwitch
            disabled={options.jsonData.interfacev2 ?? false}
            value={options.jsonData.interfacev2 ?? false}
            onChange={(e) =>
              onOptionsChange({ ...options, jsonData: { ...options.jsonData, interfacev2: e.currentTarget.checked } })
            }
          />
        </InlineField>
      </InlineFieldRow>

      <Divider spacing={4} />
    </>
  );
}
