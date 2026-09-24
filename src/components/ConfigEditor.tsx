import React from 'react';
import { Divider, InlineField, InlineFieldRow, InlineSwitch } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { SkywalkingDataSourceOptions } from '../types';
import { Auth, ConnectionSettings, convertLegacyAuthProps } from '@grafana/plugin-ui';
interface Props extends DataSourcePluginOptionsEditorProps<SkywalkingDataSourceOptions> {}

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
      <Auth {...convertLegacyAuthProps({ config: options, onChange: onOptionsChange })} />
      <InlineFieldRow>
        <InlineField
          label="Interface Version v2"
          labelWidth={24}
          tooltip="Use the SkyWalking v2 trace query API (queryTraces). Requires OAP 10.3.0 or newer. If Save & test reports that your OAP server doesn't support v2, turn this switch back off to fall back to the v1 API."
        >
          <InlineSwitch
            label="Interface Version v2"
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
