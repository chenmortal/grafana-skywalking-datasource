import { DataSourcePlugin } from '@grafana/data';
import { SkywalkingDataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { SkywalkingQuery, SkywalkingDataSourceOptions } from './types';
import { initPluginTranslations } from '@grafana/i18n';
import pluginJson from 'plugin.json';

await initPluginTranslations(pluginJson.id);
export const plugin = new DataSourcePlugin<SkywalkingDataSource, SkywalkingQuery, SkywalkingDataSourceOptions>(
  SkywalkingDataSource
)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
