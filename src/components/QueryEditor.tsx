import { css } from '@emotion/css';
import React from 'react';
import { InlineField, InlineFieldRow, Input, RadioButtonGroup, Stack, useStyles2 } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { SkywalkingDataSource } from '../datasource';
import { MyDataSourceOptions, SkywalkingQuery, SkywalkingQueryType } from '../types';
import SearchForm from './SearchForm';

type Props = QueryEditorProps<SkywalkingDataSource, SkywalkingQuery, MyDataSourceOptions>;

export function QueryEditor({ datasource, query, onChange, onRunQuery }: Props) {
  const styles = useStyles2(getStyles);
  const onChangeQuery = (value: string) => {
    const nextQuery: SkywalkingQuery = { ...query, query: value };
    onChange(nextQuery);
  };
  const renderEditorBody = () => {
    switch (query.queryType) {
      case 'search':
        return <SearchForm datasource={datasource} query={query} onChange={onChange} />;
      case 'dependencyGraph':
        return null;
      default:
        return (
          <InlineFieldRow>
            <InlineField label="Trace ID" labelWidth={14} grow>
              <Input
                value={query.query || ''}
                onChange={(e) => onChangeQuery(e.currentTarget.value)}
                placeholder="Enter a Trace ID"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && e.shiftKey) {
                    e.preventDefault();
                    onRunQuery();
                  }
                }}
              />
            </InlineField>
          </InlineFieldRow>
        );
    }
  };

  return (
    <>
      <div className={styles.container}>
        <InlineFieldRow>
          <InlineField label="Query Type" grow={true}>
            <Stack gap={1} alignItems="center" justifyContent="space-between">
              <RadioButtonGroup<SkywalkingQueryType>
                value={query.queryType}
                options={[
                  { label: 'Search', value: 'search' },
                  { value: undefined, label: 'TraceID' },
                  { label: 'Dependency Graph', value: 'dependencyGraph' },
                ]}
                onChange={(v) =>
                  onChange({
                    ...query,
                    queryType: v,
                  })
                }
              />
            </Stack>
          </InlineField>
        </InlineFieldRow>
        {renderEditorBody()}
      </div>
    </>
  );
}

const getStyles = () => ({
  container: css({
    width: '100%',
  }),
});
