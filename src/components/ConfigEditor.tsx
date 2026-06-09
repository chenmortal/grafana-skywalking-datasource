import React  from 'react';
import { Divider} from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';
import {
  ConnectionSettings,
} from '@grafana/plugin-ui';
interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  return (
    <>
        <Divider spacing={4} />
        <ConnectionSettings config={options} onChange={onOptionsChange} urlPlaceholder="http://skywalking.example.com/graphql" />
        <Divider spacing={4} />
    </>
  );
}
