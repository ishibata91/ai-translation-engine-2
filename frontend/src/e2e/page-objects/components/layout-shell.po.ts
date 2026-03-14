import {Page} from '@playwright/test';
import {BasePO} from '../base.po';
import {HeaderPO} from './header.po';
import {SidebarPO} from './sidebar.po';

export class LayoutShellPO extends BasePO {
  private readonly header: HeaderPO;
  private readonly sidebar: SidebarPO;

  constructor(page: Page) {
    super(page);
    this.header = new HeaderPO(page);
    this.sidebar = new SidebarPO(page);
  }

  async expectVisible(): Promise<void> {
    await this.header.expectVisible();
    await this.sidebar.expectVisible();
    await this.expectNoRuntimeErrors();
  }

  async openDictionary(): Promise<void> {
    await this.sidebar.openDictionary();
  }

  async openMasterPersona(): Promise<void> {
    await this.sidebar.openMasterPersona();
  }
}
