import {test} from './fixtures/app.fixture';

test('アプリ起動時に主要レイアウトが表示される', async ({app}) => {
  await app.dashboard.open();
  await app.layoutShell.expectVisible();
});

test('ダッシュボードが初期表示される', async ({app}) => {
  await app.dashboard.open();
  await app.dashboard.expectVisible();
});

test('サイドバーから辞書構築へ遷移できる', async ({app}) => {
  await app.dashboard.open();
  await app.layoutShell.openDictionary();
  await app.dictionaryBuilder.expectVisible();
});

test('サイドバーからマスターペルソナ構築へ遷移できる', async ({app}) => {
  await app.dashboard.open();
  await app.layoutShell.openMasterPersona();
  await app.masterPersona.expectVisible();
});
