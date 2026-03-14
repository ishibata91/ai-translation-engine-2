import {expect, Page} from '@playwright/test';
import {BasePO} from '../base.po';

export class MasterPersonaPO extends BasePO {
  constructor(page: Page) {
    super(page);
  }

  async expectVisible(): Promise<void> {
    await expect(this.page.getByText('マスターペルソナ構築 (Master Persona Builder)')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }
}
