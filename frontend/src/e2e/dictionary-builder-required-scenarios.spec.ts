import {test} from './fixtures/app.fixture';
import {
  DICTIONARY_BUILDER_REQUIRED_CROSS_SEARCH_MARKERS,
  DICTIONARY_BUILDER_REQUIRED_ENTRY_MARKERS,
  DICTIONARY_BUILDER_REQUIRED_QUERY,
  DICTIONARY_BUILDER_REQUIRED_SOURCE_FILES,
} from './fixtures/dictionary-builder/mock-data';

const PRIMARY_SOURCE = DICTIONARY_BUILDER_REQUIRED_SOURCE_FILES[0];

test('DictionaryBuilder の必須シナリオ: 一覧表示', async ({app}) => {
  await app.dictionaryBuilder.open();
  await app.dictionaryBuilder.expectVisible();
  await app.dictionaryBuilder.expectSourceList(DICTIONARY_BUILDER_REQUIRED_SOURCE_FILES);
});

test('DictionaryBuilder の必須シナリオ: 詳細確認から編集画面へ遷移', async ({app}) => {
  await app.dictionaryBuilder.open();
  await app.dictionaryBuilder.selectSource(PRIMARY_SOURCE);
  await app.dictionaryBuilder.expectDetailPane(PRIMARY_SOURCE);
  await app.dictionaryBuilder.openEntriesEditor();
  await app.dictionaryBuilder.expectEntriesEditor(PRIMARY_SOURCE, DICTIONARY_BUILDER_REQUIRED_ENTRY_MARKERS);
});

test('DictionaryBuilder の必須シナリオ: 横断検索結果へ遷移', async ({app}) => {
  await app.dictionaryBuilder.open();
  await app.dictionaryBuilder.openCrossSearch();
  await app.dictionaryBuilder.searchCrossEntries(DICTIONARY_BUILDER_REQUIRED_QUERY);
  await app.dictionaryBuilder.expectCrossSearchResults(
    DICTIONARY_BUILDER_REQUIRED_QUERY,
    DICTIONARY_BUILDER_REQUIRED_CROSS_SEARCH_MARKERS,
  );
});
