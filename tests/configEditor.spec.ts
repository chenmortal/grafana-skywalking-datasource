import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByRole('textbox', { name: 'Data source connection URL' })).toBeVisible();
  await expect(page.getByText('Interface Version', { exact: false })).toBeVisible();
});

test('"Save & test" should fail when no valid URL is configured', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error');
});

test('"Interface Version v2" switch can be turned back off (recovery flow)', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  const v2Switch = page.getByRole('switch', { name: /Interface Version v2/i });
  await v2Switch.check({ force: true });
  await expect(v2Switch).toBeChecked();
  // The switch must not lock itself once enabled: when the OAP backend turns
  // out not to support queryTracesV2, "Save & test" tells users to turn it
  // off, so it has to stay operable.
  await expect(v2Switch).toBeEnabled();
  await v2Switch.uncheck({ force: true });
  await expect(v2Switch).not.toBeChecked();
});
