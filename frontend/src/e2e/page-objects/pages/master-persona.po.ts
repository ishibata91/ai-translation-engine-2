import {expect, Page} from '@playwright/test';
import {APP_ROUTES} from '../../helpers/routes';
import {BasePO} from '../base.po';

type PersonaDetailExpectation = {
  formId: string;
  name: string;
};

type ModelSettingsExpectation = {
  model?: string;
  provider?: string;
  syncConcurrency?: string;
  temperature?: string;
};

export class MasterPersonaPO extends BasePO {
  constructor(page: Page) {
    super(page);
  }

  async open(): Promise<void> {
    await this.page.goto(APP_ROUTES.masterPersona);
    await this.page.waitForLoadState('domcontentloaded');
    await this.waitForHash(APP_ROUTES.masterPersona);
    await this.expectNoRuntimeErrors();
  }

  async expectVisible(): Promise<void> {
    await expect(this.page.getByText('マスターペルソナ構築 (Master Persona Builder)')).toBeVisible();
    await expect(this.page.getByRole('button', {name: 'JSON選択'})).toBeVisible();
    await expect(this.page.getByText('全体進捗')).toBeVisible();
    await expect(this.page.getByRole('heading', {name: 'ユーザープロンプト'})).toBeVisible();
    await expect(this.page.getByRole('heading', {name: 'システムプロンプト'})).toBeVisible();
    await expect(this.page.getByText('ペルソナ生成モデル設定')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectNpcList(records: string[]): Promise<void> {
    for (const record of records) {
      await expect(this.page.getByRole('cell', {name: record}).first()).toBeVisible();
    }
    await this.expectNoRuntimeErrors();
  }

  async selectNpc(nameOrFormId: string): Promise<void> {
    const row = this.page.locator('tbody tr', {hasText: nameOrFormId}).first();
    await expect(row).toBeVisible();
    await row.click();

    const emptyDetail = this.page.getByText('表示するNPCを選択してください');
    if (await emptyDetail.isVisible()) {
      await row.click();
    }

    await this.expectNoRuntimeErrors();
  }

  async expectPersonaDetail(record: PersonaDetailExpectation): Promise<void> {
    const detailHeading = this.page.locator('h3').filter({hasText: record.name}).first();
    await expect(detailHeading).toBeVisible();
    await expect(detailHeading).toContainText(record.formId);
    await expect(this.page.getByText('生成されたペルソナ情報 (プロンプト動的注入用)')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async expectPromptCards(): Promise<void> {
    await expect(this.page.getByRole('heading', {name: 'ユーザープロンプト'})).toBeVisible();
    await expect(this.page.getByRole('heading', {name: 'システムプロンプト'})).toBeVisible();
    await expect(this.page.getByText('編集可能')).toBeVisible();
    await expect(this.page.getByText('Read Only')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }

  async editUserPrompt(value: string): Promise<void> {
    const userPromptCard = this.page.locator('div.card', {has: this.page.getByRole('heading', {name: 'ユーザープロンプト'})}).first();
    const textarea = userPromptCard.locator('textarea');
    await expect(textarea).toBeEditable();
    await textarea.fill(value);
    await this.expectNoRuntimeErrors();
  }

  async expectUserPromptValue(value: string): Promise<void> {
    const userPromptCard = this.page.locator('div.card', {has: this.page.getByRole('heading', {name: 'ユーザープロンプト'})}).first();
    await expect(userPromptCard.locator('textarea')).toHaveValue(value);
    await this.expectNoRuntimeErrors();
  }

  async expectSystemPromptReadonly(): Promise<void> {
    const systemPromptCard = this.page.locator('div.card', {has: this.page.getByRole('heading', {name: 'システムプロンプト'})}).first();
    const textarea = systemPromptCard.locator('textarea');
    await expect(textarea).toHaveAttribute('readonly', '');
    await this.expectNoRuntimeErrors();
  }

  async changeProvider(provider: string): Promise<void> {
    const providerSelect = this.page.locator('details:has-text("ペルソナ生成モデル設定") select').first();
    await providerSelect.selectOption(provider);
    await this.expectNoRuntimeErrors();
  }

  async changeTemperature(value: string): Promise<void> {
    const modelSection = this.page.locator('details:has-text("ペルソナ生成モデル設定")');
    const temperatureSlider = modelSection.locator('input[type="range"]').nth(1);
    await temperatureSlider.fill(value);
    await temperatureSlider.press('Enter');
    await this.expectNoRuntimeErrors();
  }

  async expectModelSettingsValue(expected: ModelSettingsExpectation): Promise<void> {
    const modelSection = this.page.locator('details:has-text("ペルソナ生成モデル設定")');

    if (expected.provider != null) {
      const providerSelect = modelSection.locator('select').first();
      await expect(providerSelect).toHaveValue(expected.provider);
    }

    if (expected.model != null) {
      const modelSelect = modelSection.locator('select').nth(1);
      await expect(modelSelect).toHaveValue(expected.model);
    }

    if (expected.syncConcurrency != null) {
      await expect(modelSection.getByText(expected.syncConcurrency, {exact: false})).toBeVisible();
    }

    if (expected.temperature != null) {
      await expect(modelSection.getByText(expected.temperature, {exact: false})).toBeVisible();
    }

    await this.expectNoRuntimeErrors();
  }

  async chooseJson(): Promise<void> {
    await this.page.getByRole('button', {name: 'JSON選択'}).click();
    await this.expectNoRuntimeErrors();
  }

  async expectStartReady(jsonPath: string): Promise<void> {
    await expect(this.page.locator('input[readonly]').first()).toHaveValue(jsonPath);
    await expect(this.page.getByRole('button', {name: '新規タスク開始'})).toBeEnabled();
    await this.expectNoRuntimeErrors();
  }

  async startTask(): Promise<void> {
    await this.page.getByRole('button', {name: '新規タスク開始'}).click();
    await this.expectNoRuntimeErrors();
  }

  async expectTaskStarted(): Promise<void> {
    await expect(this.page.getByRole('button', {name: '生成中...'})).toBeVisible();
    await expect(this.page.getByText('Job: MasterPersonaGeneration (Running)')).toBeVisible();
    await this.expectNoRuntimeErrors();
  }
}
