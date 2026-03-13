import {expect, Page} from '@playwright/test';

export const APP_ROUTES = {
  dashboard: '/#/',
  dictionary: '/#/dictionary',
  masterPersona: '/#/master_persona',
} as const;

export async function expectCurrentHash(page: Page, expectedHashPath: string): Promise<void> {
  await expect
    .poll(async () => page.evaluate(() => window.location.hash))
    .toBe(expectedHashPath.replace('/#', '#'));
}
