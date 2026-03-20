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

  private terminologyPanel(): Locator {
    return this.page
      .locator('.tab-content-panel')
      .filter({has: this.page.getByRole('heading', {name: '単語翻訳 phase'})})
      .first();
  }

  private terminologyModelSettings(): Locator {
    return this.terminologyPanel()
      .locator('.card, .collapse')
      .filter({has: this.page.getByText('単語翻訳モデル設定')})
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

  async open(scenario?: string): Promise<void> {
    const route = scenario
      ? `${APP_ROUTES.translationFlow}?tfScenario=${encodeURIComponent(scenario)}`
      : APP_ROUTES.translationFlow;
    await this.page.goto(route);
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

  async proceedToTerminologyPhase(): Promise<void> {
    await this.loadPanel().getByRole('button', {name: 'ロード完了して次へ'}).click();
    await expect(this.page.getByRole('tab', {name: '単語翻訳'})).toHaveClass(/tab-active/);
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyPhaseVisible(): Promise<void> {
    const panel = this.terminologyPanel();
    await expect(panel).toBeVisible();
    await expect(panel.getByRole('heading', {name: '対象単語リスト', exact: true})).toBeVisible();
    await expect(panel.getByRole('columnheader', {name: 'Translated Text'})).toBeVisible();
    await expect(panel.getByText('保存件数')).toBeVisible();
    await expect(panel.getByText('失敗件数')).toBeVisible();
    await expect(panel.getByRole('button', {name: '単語翻訳を実行'})).toBeVisible();
    await expect(this.terminologyModelSettings()).toBeVisible();
    await expect(this.terminologyModelSettings().locator('select').nth(1)).toHaveValue('local-terminology-model');
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyTargetVisible(recordType: string, editorId: string, sourceText: string, variant: string, sourceFile: string): Promise<void> {
    const panel = this.terminologyPanel();
    await expect(panel.getByRole('cell', {name: recordType}).first()).toBeVisible();
    await expect(panel.getByRole('cell', {name: editorId}).first()).toBeVisible();
    await expect(panel.getByRole('cell', {name: sourceText}).first()).toBeVisible();
    await expect(panel.getByRole('cell', {name: variant}).first()).toBeVisible();
    await expect(panel.getByRole('cell', {name: sourceFile}).first()).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyUntranslatedBadge(): Promise<void> {
    await expect(this.terminologyPanel().getByText('未翻訳').first()).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologySystemPromptReadOnly(): Promise<void> {
    const textarea = this.terminologyPanel()
      .locator('.card')
      .filter({has: this.page.getByRole('heading', {name: 'System Prompt'})})
      .locator('textarea');
    await expect(textarea).toHaveAttribute('readonly', '');
    await this.expectNoRuntimeErrors();
  }

  async runTerminologyPhase(): Promise<void> {
    const runButton = this.terminologyPanel().getByRole('button', {name: '単語翻訳を実行'});
    await expect(runButton).toBeEnabled();
    await runButton.click();
    await expect(this.terminologyPanel().getByText('単語翻訳完了')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async startTerminologyPhase(): Promise<void> {
    const runButton = this.terminologyPanel().getByRole('button', {name: '単語翻訳を実行'});
    await expect(runButton).toBeEnabled();
    await runButton.click();
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyRunningProgress(progressLabel: string): Promise<void> {
    const panel = this.terminologyPanel();
    await expect(panel.getByRole('button', {name: '単語翻訳を実行中...'})).toBeDisabled();
    await expect(panel.getByText(progressLabel)).toBeVisible();
    await expect(panel.getByText('読込中')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyControlsDisabled(): Promise<void> {
    const panel = this.terminologyPanel();
    await expect(panel.getByRole('button', {name: '状態を再読込'})).toBeDisabled();
    await expect(panel.getByRole('button', {name: '次へ'})).toBeDisabled();
    await expect(panel.getByRole('button', {name: '単語翻訳を確定して次へ'})).toBeDisabled();
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyStatus(text: string): Promise<void> {
    await expect(this.terminologyPanel().getByText(text)).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyEmptyState(): Promise<void> {
    await expect(this.terminologyPanel().getByText('ロード済みデータに Terminology 対象 REC がありません。')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyErrorPanel(message: string): Promise<void> {
    await expect(this.terminologyPanel().getByText(message)).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyAdvanceEnabled(): Promise<void> {
    await expect(this.terminologyPanel().getByRole('button', {name: '単語翻訳を確定して次へ'})).toBeEnabled();
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologySummary(savedCount: string, failedCount: string): Promise<void> {
    const panel = this.terminologyPanel();
    await expect(panel.getByText('保存件数').locator('..')).toContainText(savedCount);
    await expect(panel.getByText('失敗件数').locator('..')).toContainText(failedCount);
    await this.expectNoRuntimeErrors();
  }

  async expectTerminologyTranslatedText(text: string): Promise<void> {
    await expect(this.terminologyPanel().getByText(text, {exact: false}).first()).toBeVisible();
    await this.expectNoRuntimeErrors();
  }
}
