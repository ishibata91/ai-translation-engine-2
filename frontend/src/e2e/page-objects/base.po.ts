import {expect, Page} from '@playwright/test';

type RuntimeErrorStore = {
  consoleErrors: string[];
  pageErrors: string[];
};

const runtimeErrorsByPage = new WeakMap<Page, RuntimeErrorStore>();

function getRuntimeErrorStore(page: Page): RuntimeErrorStore {
  const existing = runtimeErrorsByPage.get(page);
  if (existing) {
    return existing;
  }

  const created: RuntimeErrorStore = {
    consoleErrors: [],
    pageErrors: [],
  };

  page.on('pageerror', (error) => {
    created.pageErrors.push(error.message);
  });

  page.on('console', (message) => {
    if (message.type() === 'error') {
      created.consoleErrors.push(message.text());
    }
  });

  runtimeErrorsByPage.set(page, created);
  return created;
}

export class BasePO {
  protected readonly page: Page;

  protected constructor(page: Page) {
    this.page = page;
    getRuntimeErrorStore(page);
  }

  protected async waitForHash(expectedHashPath: string): Promise<void> {
    await expect
      .poll(async () => this.page.evaluate(() => window.location.hash))
      .toBe(expectedHashPath.replace('/#', '#'));
  }

  protected async expectNoRuntimeErrors(): Promise<void> {
    const errors = getRuntimeErrorStore(this.page);

    if (errors.pageErrors.length === 0 && errors.consoleErrors.length === 0) {
      return;
    }

    const details = [
      ...errors.pageErrors.map((entry) => `pageerror: ${entry}`),
      ...errors.consoleErrors.map((entry) => `console: ${entry}`),
    ].join('\n');

    throw new Error(`アプリ初期化時に例外を検知しました\n${details}`);
  }
}
