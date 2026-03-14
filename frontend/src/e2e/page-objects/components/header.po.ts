import {expect, Page} from '@playwright/test';
import {BasePO} from '../base.po';

export class HeaderPO extends BasePO {
  constructor(page: Page) {
    super(page);
  }

  async expectVisible(): Promise<void> {
    await expect(this.page.getByText('AI Translation Engine 2')).toBeVisible();
  }
}
