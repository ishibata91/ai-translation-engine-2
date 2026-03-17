import {test} from './fixtures/app.fixture';
import {
    TRANSLATION_FLOW_EXPECTED_FILE_NAMES,
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
  await app.translationFlow.expectTerminologySystemPromptReadOnly();
  await app.translationFlow.runTerminologyPhase();
  await app.translationFlow.expectTerminologySummary('8', '8', '0');
});
