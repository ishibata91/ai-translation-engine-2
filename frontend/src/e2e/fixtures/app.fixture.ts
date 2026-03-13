import {expect, Page, test as base} from '@playwright/test';
import {APP_ROUTES, expectCurrentHash} from '../helpers/routes';
import {installWailsMocks} from '../helpers/wails-mock';

type AppHarness = {
  openDashboard: () => Promise<void>;
  expectShellVisible: () => Promise<void>;
  expectDashboardVisible: () => Promise<void>;
  openDictionaryBySidebar: () => Promise<void>;
  openMasterPersonaBySidebar: () => Promise<void>;
  expectDictionaryVisible: () => Promise<void>;
  expectMasterPersonaVisible: () => Promise<void>;
};

function createHarness(page: Page): AppHarness {
  const pageErrors: string[] = [];
  const consoleErrors: string[] = [];
  page.on('pageerror', (error) => {
    pageErrors.push(error.message);
  });
  page.on('console', (message) => {
    if (message.type() === 'error') {
      consoleErrors.push(message.text());
    }
  });

  const assertNoRuntimeError = async (): Promise<void> => {
    if (pageErrors.length === 0 && consoleErrors.length === 0) {
      return;
    }

    const details = [
      ...pageErrors.map((entry) => `pageerror: ${entry}`),
      ...consoleErrors.map((entry) => `console: ${entry}`),
    ].join('\n');

    throw new Error(`アプリ初期化時に例外を検知しました\n${details}`);
  };

  return {
    openDashboard: async () => {
      await page.goto(APP_ROUTES.dashboard);
      await page.waitForLoadState('domcontentloaded');
      await assertNoRuntimeError();
      await expectCurrentHash(page, APP_ROUTES.dashboard);
    },
    expectShellVisible: async () => {
      await expect(page.getByText('AI Translation Engine 2')).toBeVisible();
      await expect(page.getByRole('link', { name: 'ダッシュボード' })).toBeVisible();
      await expect(page.getByRole('link', { name: '辞書構築' })).toBeVisible();
      await expect(page.getByRole('link', { name: 'マスターペルソナ構築' })).toBeVisible();
    },
    expectDashboardVisible: async () => {
      await expect(page.getByText('ダッシュボード (Dashboard)')).toBeVisible();
    },
    openDictionaryBySidebar: async () => {
      await page.getByRole('link', { name: '辞書構築' }).click();
      await expectCurrentHash(page, APP_ROUTES.dictionary);
      await assertNoRuntimeError();
    },
    openMasterPersonaBySidebar: async () => {
      await page.getByRole('link', { name: 'マスターペルソナ構築' }).click();
      await expectCurrentHash(page, APP_ROUTES.masterPersona);
      await page.waitForTimeout(300);
      await assertNoRuntimeError();
    },
    expectDictionaryVisible: async () => {
      await expect(page.getByText('辞書構築 (Dictionary Builder)')).toBeVisible();
    },
    expectMasterPersonaVisible: async () => {
      await expect(page.getByText('マスターペルソナ構築 (Master Persona Builder)')).toBeVisible();
    },
  };
}

export const test = base.extend<{ app: AppHarness }>({
  app: async ({ page }, provideApp) => {
    await installWailsMocks(page);
    await provideApp(createHarness(page));
  },
});
