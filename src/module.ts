import { DataSourcePlugin } from '@grafana/data';
import { SkywalkingDataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { SkywalkingQuery, MyDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<SkywalkingDataSource, SkywalkingQuery, MyDataSourceOptions>(SkywalkingDataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
