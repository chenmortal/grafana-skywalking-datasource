import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor with query type selector', async ({
  panelEditPage,
  readProvisionedDataSource,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('radio', { name: 'TraceID' })).toBeVisible();
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('radio', { name: 'Search' })).toBeVisible();
});

test('should render Trace ID input in default mode', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('textbox', { name: 'Trace ID' })).toBeVisible();
});

test('should switch to search mode and hide trace ID input', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  const row = panelEditPage.getQueryEditorRow('A');
  await expect(row.getByRole('textbox', { name: 'Trace ID' })).toBeVisible();
  await row.getByRole('radio', { name: 'Search' }).click();
  await expect(row.getByRole('textbox', { name: 'Trace ID' })).toHaveCount(0);
});
