import {expect, Page} from '@playwright/test';
import {BasePO} from '../base.po';

export class DictionaryBuilderPO extends BasePO {
  constructor(page: Page) {
    super(page);
  }

  async expectVisible(): Promise<void> {
    await expect(this.page.getByText('辞書構築 (Dictionary Builder)')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }
}
