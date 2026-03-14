import {expect, Page} from '@playwright/test';
import {APP_ROUTES} from '../../helpers/routes';
import {BasePO} from '../base.po';

export class SidebarPO extends BasePO {
  constructor(page: Page) {
    super(page);
  }

  async expectVisible(): Promise<void> {
    await expect(this.page.getByRole('link', {name: 'ダッシュボード'})).toBeVisible();
    await expect(this.page.getByRole('link', {name: '辞書構築'})).toBeVisible();
    await expect(this.page.getByRole('link', {name: 'マスターペルソナ構築'})).toBeVisible();
  }

  async openDictionary(): Promise<void> {
    await this.page.getByRole('link', {name: '辞書構築'}).click();
    await this.waitForHash(APP_ROUTES.dictionary);
    await this.expectNoRuntimeErrors();
  }

  async openMasterPersona(): Promise<void> {
    await this.page.getByRole('link', {name: 'マスターペルソナ構築'}).click();
    await this.waitForHash(APP_ROUTES.masterPersona);
    await this.page.waitForTimeout(300);
    await this.expectNoRuntimeErrors();
  }
}
