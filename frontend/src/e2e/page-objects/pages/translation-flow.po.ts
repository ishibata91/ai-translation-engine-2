import {expect, Locator, Page} from '@playwright/test';
import {APP_ROUTES} from '../../helpers/routes';
import {BasePO} from '../base.po';

export class TranslationFlowPO extends BasePO {
  constructor(page: Page) {
    super(page);
  }

  private loadPanel(): Locator {
    return this.page
      .locator('.tab-content-panel')
      .filter({has: this.page.getByRole('heading', {name: 'データロード'})})
      .first();
  }

  private fileTable(fileName: string): Locator {
    return this.loadPanel()
      .locator('.collapse')
      .filter({hasText: fileName})
      .first();
  }

  private async ensureExpanded(fileName: string): Promise<void> {
    const table = this.fileTable(fileName);
    const content = table.locator('.collapse-content').first();
    if (await content.isVisible()) {
      return;
    }
    await table.evaluate((element) => {
      if (element instanceof HTMLDetailsElement) {
        element.open = true;
        return;
      }
      element.setAttribute('open', '');
    });
    await expect(content).toBeVisible();
  }

  private async ensureCollapsed(fileName: string): Promise<void> {
    const table = this.fileTable(fileName);
    const content = table.locator('.collapse-content').first();
    if (!(await content.isVisible())) {
      return;
    }
    await table.evaluate((element) => {
      if (element instanceof HTMLDetailsElement) {
        element.open = false;
        return;
      }
      element.removeAttribute('open');
    });
    await expect(content).not.toBeVisible();
  }

  async open(): Promise<void> {
    await this.page.goto(APP_ROUTES.translationFlow);
    await this.page.waitForLoadState('domcontentloaded');
    await this.waitForHash(APP_ROUTES.translationFlow);
    await this.expectNoRuntimeErrors();
  }

  async expectLoadPhaseVisible(): Promise<void> {
    const panel = this.loadPanel();
    await expect(this.page.getByRole('tab', {name: 'データロード'})).toHaveClass(/tab-active/);
    await expect(panel).toBeVisible();
    await expect(panel.getByRole('heading', {name: 'データロード'})).toBeVisible();
    await expect(panel.getByRole('button', {name: 'ファイルを選択'})).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async selectFiles(): Promise<void> {
    await this.loadPanel().getByRole('button', {name: 'ファイルを選択'}).click();
    await this.expectNoRuntimeErrors();
  }

  async expectSelectedFiles(fileNames: readonly string[]): Promise<void> {
    for (const fileName of fileNames) {
      const badge = this.loadPanel().locator('.badge').filter({hasText: fileName});
      await expect(badge).toBeVisible();
    }
    await this.expectNoRuntimeErrors();
  }

  async loadSelectedFiles(): Promise<void> {
    await this.loadPanel().getByRole('button', {name: 'ロード実行'}).click();
    await this.expectNoRuntimeErrors();
  }

  async expectFileTables(fileNames: readonly string[]): Promise<void> {
    for (const fileName of fileNames) {
      const table = this.fileTable(fileName);
      await expect(table).toBeVisible();
      await expect(table.getByRole('heading', {name: new RegExp(fileName)})).toBeVisible();
    }
    await this.expectNoRuntimeErrors();
  }

  async expandFileTable(fileName: string): Promise<void> {
    await this.ensureExpanded(fileName);
    await this.expectNoRuntimeErrors();
  }

  async collapseFileTable(fileName: string): Promise<void> {
    await this.ensureCollapsed(fileName);
    await this.expectNoRuntimeErrors();
  }

  async goToNextPageForFile(fileName: string): Promise<void> {
    await this.ensureExpanded(fileName);
    const nextButton = this.fileTable(fileName).getByRole('button', {name: '次へ'});
    await nextButton.evaluate((element) => {
      if (element instanceof HTMLButtonElement) {
        element.click();
      }
    });
    await this.expectNoRuntimeErrors();
  }

  async expectMarkerVisibleInFile(fileName: string, marker: string): Promise<void> {
    await this.ensureExpanded(fileName);
    await expect(this.fileTable(fileName).locator('tbody').getByText(marker, {exact: false})).toBeVisible();
    await this.expectNoRuntimeErrors();
  }
}
