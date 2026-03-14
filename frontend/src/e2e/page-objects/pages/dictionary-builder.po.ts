import {expect, Page} from '@playwright/test';
import {APP_ROUTES} from '../../helpers/routes';
import {BasePO} from '../base.po';

export class DictionaryBuilderPO extends BasePO {
  constructor(page: Page) {
    super(page);
  }

  async open(): Promise<void> {
    await this.page.goto(APP_ROUTES.dictionary);
    await this.page.waitForLoadState('domcontentloaded');
    await this.waitForHash(APP_ROUTES.dictionary);
    await this.expectNoRuntimeErrors();
  }

  async expectVisible(): Promise<void> {
    await expect(this.page.getByText('辞書構築 (Dictionary Builder)')).toBeVisible();
    await expect(this.page.getByText('XMLインポート (xTranslator形式)')).toBeVisible();
    await expect(this.page.getByText('登録済み辞書ソース一覧')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectSourceList(fileNames: string[]): Promise<void> {
    for (const fileName of fileNames) {
      await expect(this.page.getByText(fileName)).toBeVisible();
    }
    await this.expectNoRuntimeErrors();
  }

  async selectSource(fileName: string): Promise<void> {
    const targetRow = this.page.locator('tbody tr', {hasText: fileName}).first();
    await expect(targetRow).toBeVisible();
    await targetRow.click();
    await this.expectNoRuntimeErrors();
  }

  async expectDetailPane(fileName: string): Promise<void> {
    await expect(this.page.getByText(`詳細: ${fileName}`)).toBeVisible();
    await expect(this.page.getByRole('button', {name: '📋 エントリを表示・編集'})).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async openEntriesEditor(): Promise<void> {
    await this.page.getByRole('button', {name: '📋 エントリを表示・編集'}).click();
    await this.expectNoRuntimeErrors();
  }

  async expectEntriesEditor(fileName: string, markers: string[]): Promise<void> {
    await expect(this.page.getByText(`エントリ編集: ${fileName}`)).toBeVisible();
    for (const marker of markers) {
      await expect(this.page.getByText(marker)).toBeVisible();
    }
    await this.expectNoRuntimeErrors();
  }

  async openCrossSearch(): Promise<void> {
    await this.page.getByRole('button', {name: '🔎 横断検索'}).click();
    await expect(this.page.getByRole('heading', {name: '🔎 辞書横断検索'})).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async searchCrossEntries(query: string): Promise<void> {
    const modal = this.page.locator('dialog.modal-open');
    await modal.getByRole('textbox').fill(query);
    await modal.getByRole('button', {name: '🔎 検索実行'}).click();
    await this.expectNoRuntimeErrors();
  }

  async expectCrossSearchResults(query: string, markers: string[]): Promise<void> {
    await expect(this.page.getByText(`横断検索結果: "${query}"`, {exact: false})).toBeVisible();
    for (const marker of markers) {
      await expect(this.page.getByText(marker)).toBeVisible();
    }
    await this.expectNoRuntimeErrors();
  }
}
