import {expect, Page} from '@playwright/test';
import {APP_ROUTES} from '../../helpers/routes';
import {BasePO} from '../base.po';

export class DashboardPO extends BasePO {
  constructor(page: Page) {
    super(page);
  }

  async open(): Promise<void> {
    await this.page.goto(APP_ROUTES.dashboard);
    await this.page.waitForLoadState('domcontentloaded');
    await this.waitForHash(APP_ROUTES.dashboard);
    await this.expectNoRuntimeErrors();
  }

  async expectVisible(): Promise<void> {
    await expect(this.page.getByText('ダッシュボード (Dashboard)')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectTaskVisible(taskName: string): Promise<void> {
    await expect(this.page.getByRole('cell', {name: taskName})).toBeVisible();
  }

  async openDeleteModal(taskID: string): Promise<void> {
    await this.page.getByTestId(`dashboard-delete-${taskID}`).click();
    await expect(this.page.getByTestId('dashboard-delete-modal')).toBeVisible();
  }

  async confirmDelete(): Promise<void> {
    await this.page.getByTestId('dashboard-delete-confirm').click();
  }

  async expectTaskNotVisible(taskName: string): Promise<void> {
    await expect(this.page.getByRole('cell', {name: taskName})).toHaveCount(0);
  }
}
