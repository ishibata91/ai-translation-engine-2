import {test} from './fixtures/app.fixture';
import {
  TRANSLATION_FLOW_EXPECTED_FILE_NAMES,
  TRANSLATION_FLOW_MAIN_TRANSLATION_QUEST_LABEL,
  TRANSLATION_FLOW_MAIN_TRANSLATION_SECOND_CONVERSATION_LABEL,
  TRANSLATION_FLOW_PAGE_TWO_MARKER,
  TRANSLATION_FLOW_SECOND_FILE_MARKER,
} from './fixtures/translation-flow/mock-data';

test('TranslationFlow の必須シナリオ: 初期表示はデータロードフェーズ', async ({app}) => {
  await app.translationFlow.open();
  await app.translationFlow.expectLoadPhaseVisible();
});

test('TranslationFlow の必須シナリオ: 複数ファイルをロードしてファイル別テーブルを表示', async ({app}) => {
  await app.translationFlow.open();
  await app.translationFlow.selectFiles();
  await app.translationFlow.expectSelectedFiles(TRANSLATION_FLOW_EXPECTED_FILE_NAMES);
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.expectFileTables(TRANSLATION_FLOW_EXPECTED_FILE_NAMES);
  await app.translationFlow.expectMarkerVisibleInFile(
    TRANSLATION_FLOW_EXPECTED_FILE_NAMES[1],
    TRANSLATION_FLOW_SECOND_FILE_MARKER,
  );
});

test('TranslationFlow の必須シナリオ: ファイル単位で折りたたみとページ切り替え', async ({app}) => {
  await app.translationFlow.open();
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();

  await app.translationFlow.collapseFileTable(TRANSLATION_FLOW_EXPECTED_FILE_NAMES[0]);
  await app.translationFlow.expectMarkerVisibleInFile(
    TRANSLATION_FLOW_EXPECTED_FILE_NAMES[1],
    TRANSLATION_FLOW_SECOND_FILE_MARKER,
  );

  await app.translationFlow.expandFileTable(TRANSLATION_FLOW_EXPECTED_FILE_NAMES[0]);
  await app.translationFlow.goToNextPageForFile(TRANSLATION_FLOW_EXPECTED_FILE_NAMES[0]);
  await app.translationFlow.expectMarkerVisibleInFile(
    TRANSLATION_FLOW_EXPECTED_FILE_NAMES[0],
    TRANSLATION_FLOW_PAGE_TWO_MARKER,
  );
});

test('TranslationFlow の必須シナリオ: 単語翻訳 phase の設定復元と実行サマリ表示', async ({app}) => {
  await app.translationFlow.open();
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.expectTerminologyPhaseVisible();
  await app.translationFlow.expectTerminologyTargetVisible(
    'NPC_:FULL',
    'NPC_B_03',
    'NPC Name B-03',
    'full',
    'Update.esm.extract.json',
  );
  await app.translationFlow.expectTerminologyUntranslatedBadge();
  await app.translationFlow.expectTerminologySystemPromptReadOnly();
  await app.translationFlow.runTerminologyPhase();
  await app.translationFlow.expectTerminologySummary('8', '0');
  await app.translationFlow.expectTerminologyTranslatedText('NPC 名 B-03');
});

test('TranslationFlow の必須シナリオ: 単語翻訳実行中は progress と loading と操作無効化を表示', async ({app}) => {
  await app.translationFlow.open();
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.startTerminologyPhase();
  await app.translationFlow.expectTerminologyRunningProgress('2 / 8 件を処理中');
  await app.translationFlow.expectTerminologyControlsDisabled();
  await app.translationFlow.expectTerminologySummary('8', '0');
});

test('TranslationFlow の必須シナリオ: partial completion では未翻訳行を残したまま次へ進める', async ({app}) => {
  await app.translationFlow.open('partial');
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.startTerminologyPhase();
  await app.translationFlow.expectTerminologyStatus('単語翻訳完了（一部失敗あり）');
  await app.translationFlow.expectTerminologySummary('7', '1');
  await app.translationFlow.expectTerminologyUntranslatedBadge();
  await app.translationFlow.expectTerminologyAdvanceEnabled();
});

test('TranslationFlow の必須シナリオ: terminology 対象 0 件では empty state を表示する', async ({app}) => {
  await app.translationFlow.open('empty');
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.expectTerminologyPhaseVisible();
  await app.translationFlow.expectTerminologyEmptyState();
});

test('TranslationFlow の必須シナリオ: terminology 対象一覧取得失敗では error panel を表示する', async ({app}) => {
  await app.translationFlow.open('error');
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.expectTerminologyErrorPanel('対象単語リストの取得に失敗しました');
});

test('TranslationFlow の必須シナリオ: 既存 task 再表示で terminology の結果を復元する', async ({app}) => {
  await app.translationFlow.open('resume');
  await app.translationFlow.expectLoadPhaseVisible();
  await app.translationFlow.expectFileTables(TRANSLATION_FLOW_EXPECTED_FILE_NAMES);
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.expectTerminologyPhaseVisible();
  await app.translationFlow.expectTerminologySummary('8', '0');
  await app.translationFlow.expectTerminologyTranslatedText('NPC 名 B-03');
});

test('TranslationFlow の必須シナリオ: 本文翻訳 phase の基本表示とロックを観測できる', async ({app}) => {
  await app.translationFlow.open();
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.runTerminologyPhase();
  await app.translationFlow.proceedToPersonaPhase();
  await app.translationFlow.runPersonaPhase();
  await app.translationFlow.proceedToTranslationPhase();
  await app.translationFlow.expectTranslationPhaseVisible();
  await app.translationFlow.startMainTranslationAndExpectLock();
  await app.translationFlow.expectMainTranslationRowStatus('AI翻訳済み');
});

test('TranslationFlow の必須シナリオ: 本文翻訳で dirty warning を表示できる', async ({app}) => {
  await app.translationFlow.open();
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.runTerminologyPhase();
  await app.translationFlow.proceedToPersonaPhase();
  await app.translationFlow.runPersonaPhase();
  await app.translationFlow.proceedToTranslationPhase();
  await app.translationFlow.runMainTranslation();
  await app.translationFlow.editMainTranslationDraft('手修正テキスト');
  await app.translationFlow.selectMainTranslationRow(TRANSLATION_FLOW_MAIN_TRANSLATION_SECOND_CONVERSATION_LABEL);
  await app.translationFlow.expectDirtyWarning();
  await app.translationFlow.dismissDirtyWarning();
});

test('TranslationFlow の必須シナリオ: partial failed で未翻訳 next warning を表示できる', async ({app}) => {
  await app.translationFlow.open('main-partial');
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.runTerminologyPhase();
  await app.translationFlow.proceedToPersonaPhase();
  await app.translationFlow.runPersonaPhase();
  await app.translationFlow.proceedToTranslationPhase();
  await app.translationFlow.runMainTranslation();
  await app.translationFlow.proceedFromMainTranslation();
  await app.translationFlow.expectNextWarning(1);
});

test('TranslationFlow の必須シナリオ: full failed では次へを無効化する', async ({app}) => {
  await app.translationFlow.open('main-fullfailed');
  await app.translationFlow.selectFiles();
  await app.translationFlow.loadSelectedFiles();
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.runTerminologyPhase();
  await app.translationFlow.proceedToPersonaPhase();
  await app.translationFlow.runPersonaPhase();
  await app.translationFlow.proceedToTranslationPhase();
  await app.translationFlow.runMainTranslation();
  await app.translationFlow.expectMainTranslationRowStatus('失敗');
});

test('TranslationFlow の必須シナリオ: main translation resume でカテゴリと選択行を復元する', async ({app}) => {
  await app.translationFlow.open('main-resume');
  await app.translationFlow.expectLoadPhaseVisible();
  await app.translationFlow.expectFileTables(TRANSLATION_FLOW_EXPECTED_FILE_NAMES);
  await app.translationFlow.proceedToTerminologyPhase();
  await app.translationFlow.expectTerminologySummary('8', '0');
  await app.translationFlow.proceedToPersonaPhase();
  await app.translationFlow.proceedToTranslationPhase();
  await app.translationFlow.expectTranslationPhaseVisible();
  await app.translationFlow.expectMainTranslationCategoryActive('クエスト');
  await app.translationFlow.expectMainTranslationRowSelected(TRANSLATION_FLOW_MAIN_TRANSLATION_QUEST_LABEL);
});
